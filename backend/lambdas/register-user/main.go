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
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
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

	// Password validation
	if len(req.Password) < 6 {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Password must be at least 6 characters"})
	}

	// Initialize repository
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Check if user already exists
	var existingUser models.User
	err = userRepo.GetUserByEmail(ctx, req.Email, &existingUser)
	if err == nil {
		// User exists
		return response(http.StatusConflict, ErrorResponse{Error: "User with this email already exists"})
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to process password"})
	}

	// Create user
	user := models.NewUser(req.Email, hashedPassword)
	if err := userRepo.CreateUser(ctx, user); err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
	}

	// Generate JWT token
	userID := strings.TrimPrefix(user.PK, "USER#")
	token, err := auth.GenerateToken(userID, user.Email, "user")
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
	}

	// Return success response
	return response(http.StatusCreated, RegisterResponse{
		Message: "User registered successfully",
		UserID:  userID,
		Token:   token,
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
