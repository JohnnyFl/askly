package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type GetMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

type MessageResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	ChannelID string                 `json:"channel_id"`
	Question  string                 `json:"question"`
	Answer    string                 `json:"answer"`
	Sources   []models.MessageSource `json:"sources"`
	CreatedAt string                 `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify JWT token
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	token, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		log.Printf("ERROR: Failed to extract token: %v", err)
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Missing or invalid authorization header"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		log.Printf("ERROR: Token validation failed: %v", err)
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
	}

	// Only users can get messages
	if !claims.IsUser() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
	}

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Printf("ERROR: Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	log.Printf("INFO: Fetching messages for channel: %s, user: %s", channelID, claims.UserID)

	// Add CHANNEL# prefix for DynamoDB query
	channelID = "CHANNEL#" + channelID

	// Initialize repository
	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize message repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	// Get messages for this user and channel
	var messages []*models.ChatMessage
	if err := messageRepo.ListMessagesByChannel(ctx, channelID, &messages); err != nil {
		log.Printf("ERROR: Failed to fetch messages from DynamoDB for channel %s: %v", channelID, err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to fetch messages: %v", err)})
	}

	log.Printf("INFO: Found %d total messages for channel %s", len(messages), channelID)

	// Filter messages by user (since we query by channel, need to filter)
	var userMessages []MessageResponse
	userIDWithPrefix := "USER#" + claims.UserID
	for _, msg := range messages {
		if msg.UserID == userIDWithPrefix || msg.UserID == claims.UserID {
			userMessages = append(userMessages, MessageResponse{
				ID:        strings.TrimPrefix(msg.PK, "MESSAGE#"),
				UserID:    msg.UserID,
				ChannelID: msg.ChannelID,
				Question:  msg.Question,
				Answer:    msg.Answer,
				Sources:   msg.Sources,
				CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	log.Printf("INFO: Returning %d messages for user %s", len(userMessages), claims.UserID)

	return response(http.StatusOK, GetMessagesResponse{
		Messages: userMessages,
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
