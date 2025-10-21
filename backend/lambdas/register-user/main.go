package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Message      string `json:"message"`
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid email format"})
	}

	// Password validation (Cognito requires min 8 chars with uppercase, lowercase, number)
	if len(req.Password) < 8 {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters"})
	}

	// Initialize Cognito client
	cognitoClient, err := auth.NewCognitoClient(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Register user in Cognito
	cognitoUserID, err := cognitoClient.SignUpUser(ctx, req.Email, req.Password)
	if err != nil {
		// Check for specific Cognito errors
		errMsg := err.Error()
		if strings.Contains(errMsg, "UsernameExistsException") {
			return response(http.StatusConflict, ErrorResponse{Error: "User with this email already exists"})
		}
		if strings.Contains(errMsg, "InvalidPasswordException") {
			return response(http.StatusBadRequest, ErrorResponse{Error: "Password does not meet requirements (min 8 chars, uppercase, lowercase, number)"})
		}
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
	}

	// Auto-confirm user (for POC - remove in production)
	if err := cognitoClient.ConfirmSignUp(ctx, req.Email); err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to confirm user"})
	}

	// Initialize repository
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Create user record in DynamoDB with Cognito user ID
	user := models.NewUser(req.Email, cognitoUserID) // Store Cognito ID instead of password hash
	if err := userRepo.CreateUser(ctx, user); err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user record"})
	}

	// Authenticate user to get tokens
	authResult, err := cognitoClient.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "User registered but failed to authenticate"})
	}

	// Return success response with Cognito tokens
	return response(http.StatusCreated, RegisterResponse{
		Message:      "User registered successfully",
		UserID:       cognitoUserID,
		AccessToken:  authResult.AccessToken,
		IdToken:      authResult.IdToken,
		RefreshToken: authResult.RefreshToken,
		ExpiresIn:    authResult.ExpiresIn,
	})
}

func response(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	jsonBody, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
