package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"askly/backend/shared/auth"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
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
	var req LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
	}

	// Initialize Cognito client
	cognitoClient, err := auth.NewCognitoClient(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Authenticate user with Cognito
	authResult, err := cognitoClient.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		// Check for specific Cognito errors
		errMsg := err.Error()
		if strings.Contains(errMsg, "NotAuthorizedException") || strings.Contains(errMsg, "UserNotFoundException") {
			return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		}
		if strings.Contains(errMsg, "UserNotConfirmedException") {
			return response(http.StatusForbidden, ErrorResponse{Error: "User email not confirmed"})
		}
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Authentication failed"})
	}

	// Get user info from Cognito
	cognitoUser, err := cognitoClient.GetUserFromToken(ctx, authResult.AccessToken)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user information"})
	}

	// Return success response with Cognito tokens
	return response(http.StatusOK, LoginResponse{
		Message:      "Login successful",
		UserID:       cognitoUser.Sub,
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
