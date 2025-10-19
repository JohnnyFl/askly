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

	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize user repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize transcription repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize message repository: %v", err)
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

	// Count users
	var users []*models.User
	totalUsers := 0
	if err := userRepo.ListUsers(ctx, &users); err != nil {
		log.Printf("ERROR: Failed to count users: %v", err)
	} else {
		totalUsers = len(users)
	}

	// Count transcriptions (this would need a ListTranscriptions method)
	// For now, we'll iterate through channels and count
	totalTranscriptions := 0
	for _, channel := range channels {
		var transcriptions []*models.Transcription
		if err := transcriptionRepo.ListTranscriptionsByChannel(ctx, channel.PK, &transcriptions); err != nil {
			log.Printf("WARN: Failed to count transcriptions for channel %s: %v", channel.PK, err)
		} else {
			totalTranscriptions += len(transcriptions)
		}
	}

	// Count messages (questions)
	totalQuestions := 0
	for _, user := range users {
		var messages []*models.Message
		if err := messageRepo.ListMessagesByUser(ctx, user.PK, &messages); err != nil {
			log.Printf("WARN: Failed to count messages for user %s: %v", user.PK, err)
		} else {
			totalQuestions += len(messages)
		}
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
