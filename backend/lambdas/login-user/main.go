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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Token   string `json:"token"`
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

	// Initialize repository
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Get user by email
	var user models.User
	err = userRepo.GetUserByEmail(ctx, req.Email, &user)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
	}

	// Verify password
	if !auth.IsPasswordValid(user.PasswordHash, req.Password) {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid email or password"})
	}

	// Generate JWT token
	userID := strings.TrimPrefix(user.PK, "USER#")
	token, err := auth.GenerateToken(userID, user.Email, "user")
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
	}

	// Return success response
	return response(http.StatusOK, LoginResponse{
		Message: "Login successful",
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
