package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"askly/backend/shared/auth"
	"askly/backend/shared/logger"
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
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("User registration request received")

	// Parse request body
	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		log.Warn().Str("email", req.Email).Msg("Missing required fields")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		log.Warn().Str("email", req.Email).Msg("Invalid email format")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid email format"})
	}

	// Password validation (Cognito requires min 8 chars with uppercase, lowercase, number)
	if len(req.Password) < 8 {
		log.Warn().Str("email", req.Email).Int("password_length", len(req.Password)).Msg("Password too short")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 8 characters"})
	}

	log.Debug().Str("email", req.Email).Msg("Attempting user registration in Cognito")

	// Initialize Cognito client
	cognitoClient, err := auth.NewCognitoClient(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Cognito client")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Register user in Cognito
	cognitoUserID, err := cognitoClient.SignUpUser(ctx, req.Email, req.Password)
	if err != nil {
		// Check for specific Cognito errors
		errMsg := err.Error()
		log.Error().
			Err(err).
			Str("email", req.Email).
			Str("error_type", getErrorType(errMsg)).
			Msg("Cognito user registration failed")

		if strings.Contains(errMsg, "UsernameExistsException") {
			return response(http.StatusConflict, ErrorResponse{Error: "User with this email already exists"})
		}
		if strings.Contains(errMsg, "InvalidPasswordException") {
			return response(http.StatusBadRequest, ErrorResponse{Error: "Password does not meet requirements (min 8 chars, uppercase, lowercase, number)"})
		}
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to register user"})
	}

	log.Debug().Str("email", req.Email).Str("cognito_user_id", cognitoUserID).Msg("User registered in Cognito")

	// Auto-confirm user (for POC - remove in production)
	if err := cognitoClient.ConfirmSignUp(ctx, req.Email); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to auto-confirm user")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to confirm user"})
	}

	log.Debug().Str("email", req.Email).Msg("User auto-confirmed in Cognito")

	// Initialize repository
	log.Debug().Msg("Initializing user repository")
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize user repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Create user record in DynamoDB with Cognito user ID
	user := models.NewUser(req.Email, cognitoUserID) // Store Cognito ID instead of password hash
	if err := userRepo.CreateUser(ctx, user); err != nil {
		log.Error().Err(err).Str("email", req.Email).Str("cognito_user_id", cognitoUserID).Msg("Failed to create user record in DynamoDB")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user record"})
	}

	log.Debug().Str("email", req.Email).Msg("User record created in DynamoDB")

	// Authenticate user to get tokens
	log.Debug().Str("email", req.Email).Msg("Authenticating newly registered user")
	authResult, err := cognitoClient.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to authenticate newly registered user")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "User registered but failed to authenticate"})
	}

	// Log successful registration
	duration := time.Since(start)
	log.Info().
		Str("user_id", cognitoUserID).
		Str("email", req.Email).
		Dur("duration_ms", duration).
		Msg("User registration completed successfully")

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

// getErrorType extracts the error type from Cognito error message
func getErrorType(errMsg string) string {
	if strings.Contains(errMsg, "UsernameExistsException") {
		return "UsernameExistsException"
	}
	if strings.Contains(errMsg, "InvalidPasswordException") {
		return "InvalidPasswordException"
	}
	if strings.Contains(errMsg, "InvalidParameterException") {
		return "InvalidParameterException"
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
