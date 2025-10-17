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
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Missing or invalid authorization header"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
	}

	// Only users can get messages
	if !claims.IsUser() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
	}

	// Get channel ID from query parameters
	channelID := request.QueryStringParameters["channel_id"]
	if channelID == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "channel_id query parameter is required"})
	}

	// Initialize repository
	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Get messages for this user and channel
	var messages []*models.ChatMessage
	if err := messageRepo.ListMessagesByChannel(ctx, channelID, &messages); err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch messages"})
	}

	// Filter messages by user (since we query by channel, need to filter)
	var userMessages []MessageResponse
	for _, msg := range messages {
		if msg.UserID == claims.UserID {
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
