package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"askly/backend/shared/auth"
	"askly/backend/shared/logger"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message      string `json:"message"`
	AdminID      string `json:"admin_id"`
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Admin login request received")

	// Parse request body
	var req LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		log.Warn().Str("email", req.Email).Msg("Missing required fields")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
	}

	log.Debug().Str("email", req.Email).Msg("Attempting admin authentication")

	// Initialize Cognito client
	cognitoClient, err := auth.NewCognitoClient(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cognito client")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Authenticate admin with Cognito
	authResult, err := cognitoClient.AuthenticateAdmin(ctx, req.Email, req.Password)
	if err != nil {
		// Check for specific Cognito errors
		errMsg := err.Error()
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("error_type", getErrorType(errMsg)).
			Msg("Admin authentication failed")

		if strings.Contains(errMsg, "NotAuthorizedException") || strings.Contains(errMsg, "UserNotFoundException") {
			return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
		}
		if strings.Contains(errMsg, "UserNotConfirmedException") {
			return response(http.StatusForbidden, ErrorResponse{Error: "Admin email not confirmed"})
		}
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Authentication failed"})
	}

	log.Debug().Str("email", req.Email).Msg("Cognito authentication successful")

	// Get admin info from Cognito
	cognitoUser, err := cognitoClient.GetUserFromToken(ctx, authResult.AccessToken)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to retrieve admin info from Cognito")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get admin information"})
	}

	// Log successful login
	duration := time.Since(start)
	log.Info().
		Str("admin_id", cognitoUser.Sub).
		Str("email", cognitoUser.Email).
		Dur("duration_ms", duration).
		Msg("Admin login successful")

	// Return success response with Cognito tokens
	return response(http.StatusOK, LoginResponse{
		Message:      "Admin login successful",
		AdminID:      cognitoUser.Sub,
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
