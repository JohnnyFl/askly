package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// CognitoClient wraps AWS Cognito operations
type CognitoClient struct {
	client        *cognitoidentityprovider.Client
	userPoolID    string
	clientID      string
	adminPoolID   string
	adminClientID string
}

// NewCognitoClient creates a new Cognito client
func NewCognitoClient(ctx context.Context) (*CognitoClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &CognitoClient{
		client:        cognitoidentityprovider.NewFromConfig(cfg),
		userPoolID:    os.Getenv("COGNITO_USER_POOL_ID"),
		clientID:      os.Getenv("COGNITO_CLIENT_ID"),
		adminPoolID:   os.Getenv("COGNITO_ADMIN_POOL_ID"),
		adminClientID: os.Getenv("COGNITO_ADMIN_CLIENT_ID"),
	}, nil
}

// SignUpUser registers a new user in Cognito User Pool
func (c *CognitoClient) SignUpUser(ctx context.Context, email, password string) (string, error) {
	input := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(c.clientID),
		Username: aws.String(email),
		Password: aws.String(password),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}

	output, err := c.client.SignUp(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to sign up user: %w", err)
	}

	return aws.ToString(output.UserSub), nil
}

// ConfirmSignUp confirms user registration (auto-confirm for POC)
func (c *CognitoClient) ConfirmSignUp(ctx context.Context, email string) error {
	input := &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(email),
	}

	_, err := c.client.AdminConfirmSignUp(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to confirm sign up: %w", err)
	}

	return nil
}

// AuthenticateUser authenticates a user and returns tokens
func (c *CognitoClient) AuthenticateUser(ctx context.Context, email, password string) (*AuthResult, error) {
	input := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow:   types.AuthFlowTypeAdminUserPasswordAuth,
		ClientId:   aws.String(c.clientID),
		UserPoolId: aws.String(c.userPoolID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	}

	output, err := c.client.AdminInitiateAuth(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	if output.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication failed: no tokens returned")
	}

	return &AuthResult{
		AccessToken:  aws.ToString(output.AuthenticationResult.AccessToken),
		IdToken:      aws.ToString(output.AuthenticationResult.IdToken),
		RefreshToken: aws.ToString(output.AuthenticationResult.RefreshToken),
		ExpiresIn:    output.AuthenticationResult.ExpiresIn,
	}, nil
}

// AuthenticateAdmin authenticates an admin user and returns tokens
func (c *CognitoClient) AuthenticateAdmin(ctx context.Context, email, password string) (*AuthResult, error) {
	input := &cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow:   types.AuthFlowTypeAdminUserPasswordAuth,
		ClientId:   aws.String(c.adminClientID),
		UserPoolId: aws.String(c.adminPoolID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	}

	output, err := c.client.AdminInitiateAuth(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate admin: %w", err)
	}

	if output.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication failed: no tokens returned")
	}

	return &AuthResult{
		AccessToken:  aws.ToString(output.AuthenticationResult.AccessToken),
		IdToken:      aws.ToString(output.AuthenticationResult.IdToken),
		RefreshToken: aws.ToString(output.AuthenticationResult.RefreshToken),
		ExpiresIn:    output.AuthenticationResult.ExpiresIn,
	}, nil
}

// GetUserFromToken retrieves user information from Cognito using access token
func (c *CognitoClient) GetUserFromToken(ctx context.Context, accessToken string) (*CognitoUser, error) {
	input := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(accessToken),
	}

	output, err := c.client.GetUser(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := &CognitoUser{
		Username: aws.ToString(output.Username),
	}

	// Extract email from attributes
	for _, attr := range output.UserAttributes {
		if aws.ToString(attr.Name) == "email" {
			user.Email = aws.ToString(attr.Value)
		}
		if aws.ToString(attr.Name) == "sub" {
			user.Sub = aws.ToString(attr.Value)
		}
	}

	return user, nil
}

// CreateAdminUser creates an admin user in Cognito admin pool
func (c *CognitoClient) CreateAdminUser(ctx context.Context, email, temporaryPassword string) error {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(c.adminPoolID),
		Username:          aws.String(email),
		TemporaryPassword: aws.String(temporaryPassword),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		MessageAction: types.MessageActionTypeSuppress, // Don't send welcome email
	}

	_, err := c.client.AdminCreateUser(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Set permanent password
	setPwdInput := &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(c.adminPoolID),
		Username:   aws.String(email),
		Password:   aws.String(temporaryPassword),
		Permanent:  true,
	}

	_, err = c.client.AdminSetUserPassword(ctx, setPwdInput)
	if err != nil {
		return fmt.Errorf("failed to set admin password: %w", err)
	}

	return nil
}

// AuthResult contains authentication tokens
type AuthResult struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

// CognitoUser contains user information from Cognito
type CognitoUser struct {
	Username string
	Email    string
	Sub      string // Cognito user ID
}
