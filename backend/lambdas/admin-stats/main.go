package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type StatsResponse struct {
	TotalChannels       int `json:"totalChannels"`
	TotalUsers          int `json:"totalUsers"`
	TotalQuestions      int `json:"totalQuestions"`
	TotalTranscriptions int `json:"totalTranscriptions"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify JWT token - only admins can access stats
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	token, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Missing or invalid authorization header"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
	}

	// Only admins can access stats
	if !claims.IsAdmin() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
	}

	log.Printf("INFO: Admin %s requesting stats", claims.UserID)

	// Initialize repositories
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize channel repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Count channels
	var channels []*models.Channel
	totalChannels := 0
	if err := channelRepo.ListChannels(ctx, &channels); err != nil {
		log.Printf("ERROR: Failed to count channels: %v", err)
	} else {
		totalChannels = len(channels)
	}

	// Count users by scanning the users table
	totalUsers := 0
	var users []map[string]interface{}
	db, err := repository.NewDynamoDBRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize DB: %v", err)
	} else {
		if err := db.Scan(ctx, "askly-users", &users); err != nil {
			log.Printf("ERROR: Failed to count users: %v", err)
		} else {
			totalUsers = len(users)
		}
	}

	// Count transcriptions by scanning transcriptions table
	totalTranscriptions := 0
	var transcriptions []map[string]interface{}
	if err := db.Scan(ctx, "askly-transcriptions", &transcriptions); err != nil {
		log.Printf("ERROR: Failed to count transcriptions: %v", err)
	} else {
		totalTranscriptions = len(transcriptions)
	}

	// Count messages (questions) by scanning messages table
	totalQuestions := 0
	var messages []map[string]interface{}
	if err := db.Scan(ctx, "askly-messages", &messages); err != nil {
		log.Printf("ERROR: Failed to count messages: %v", err)
	} else {
		totalQuestions = len(messages)
	}

	log.Printf("INFO: Stats - Channels: %d, Users: %d, Questions: %d, Transcriptions: %d",
		totalChannels, totalUsers, totalQuestions, totalTranscriptions)

	return response(http.StatusOK, StatsResponse{
		TotalChannels:       totalChannels,
		TotalUsers:          totalUsers,
		TotalQuestions:      totalQuestions,
		TotalTranscriptions: totalTranscriptions,
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
