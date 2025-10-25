package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"askly/backend/shared/auth"
	"askly/backend/shared/logger"
	"askly/backend/shared/response"

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
	// Handle OPTIONS preflight request
	if request.HTTPMethod == "OPTIONS" {
		return response.Options()
	}

	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("User login request received")

	// Parse request body
	var req LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response.Error(http.StatusBadRequest, "Invalid request body")
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		log.Warn().Str("email", req.Email).Msg("Missing required fields")
		return response.Error(http.StatusBadRequest, "Email and password are required")
	}

	log.Debug().Str("email", req.Email).Msg("Attempting user authentication")

	// Initialize Cognito client
	cognitoClient, err := auth.NewCognitoClient(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cognito client")
		return response.Error(http.StatusInternalServerError, "Internal server error")
	}

	// Authenticate user with Cognito
	authResult, err := cognitoClient.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		// Check for specific Cognito errors
		errMsg := err.Error()
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("error_type", getErrorType(errMsg)).
			Msg("Authentication failed")

		if strings.Contains(errMsg, "NotAuthorizedException") || strings.Contains(errMsg, "UserNotFoundException") {
			return response.Error(http.StatusUnauthorized, "Invalid email or password")
		}
		if strings.Contains(errMsg, "UserNotConfirmedException") {
			return response.Error(http.StatusForbidden, "User email not confirmed")
		}
		return response.Error(http.StatusInternalServerError, "Authentication failed")
	}

	log.Debug().Str("email", req.Email).Msg("Cognito authentication successful")

	// Get user info from Cognito
	cognitoUser, err := cognitoClient.GetUserFromToken(ctx, authResult.AccessToken)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to retrieve user info from Cognito")
		return response.Error(http.StatusInternalServerError, "Failed to get user information")
	}

	// Log successful login
	duration := time.Since(start)
	log.Info().
		Str("user_id", cognitoUser.Sub).
		Str("email", cognitoUser.Email).
		Dur("duration_ms", duration).
		Msg("User login successful")

	// Return success response with Cognito tokens
	return response.JSON(http.StatusOK, LoginResponse{
		Message:      "Login successful",
		UserID:       cognitoUser.Sub,
		AccessToken:  authResult.AccessToken,
		IdToken:      authResult.IdToken,
		RefreshToken: authResult.RefreshToken,
		ExpiresIn:    authResult.ExpiresIn,
	})
}

// getErrorType extracts the error type from Cognito error message
func getErrorType(errMsg string) string {
	if strings.Contains(errMsg, "NotAuthorizedException") {
		return "NotAuthorizedException"
	}
	if strings.Contains(errMsg, "UserNotFoundException") {
		return "UserNotFoundException"
	}
	if strings.Contains(errMsg, "UserNotConfirmedException") {
		return "UserNotConfirmedException"
	}
	return "UnknownError"
}

func main() {
	lambda.Start(handler)
}
