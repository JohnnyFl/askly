package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"askly/backend/shared/logger"
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
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Get messages request received")

	// Extract user info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	// Add user ID to logger context
	ctx = logger.WithUserID(ctx, userID)
	log = logger.FromContext(ctx)

	log.Debug().Str("email", email).Msg("User requesting messages")

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Error().Msg("Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	log.Debug().Msg("Fetching messages for user and channel")

	// Initialize repository
	log.Debug().Msg("Initializing message repository")
	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize message repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	// Get messages for this user and channel
	log.Debug().Str("channel_id", channelID).Msg("Fetching messages from database")
	var messages []*models.ChatMessage
	if err := messageRepo.ListMessagesByChannel(ctx, channelID, &messages); err != nil {
		log.Error().Err(err).Str("channel_id", channelID).Msg("Failed to fetch messages from DynamoDB")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to fetch messages: %v", err)})
	}

	log.Debug().Int("total_messages", len(messages)).Msg("Messages fetched from database")

	// Filter messages by user (since we query by channel, need to filter)
	log.Debug().Msg("Filtering messages for current user")
	var userMessages []MessageResponse
	for _, msg := range messages {
		if msg.UserID == userID {
			userMessages = append(userMessages, MessageResponse{
				ID:        msg.PK,
				UserID:    msg.UserID,
				ChannelID: msg.ChannelID,
				Question:  msg.Question,
				Answer:    msg.Answer,
				Sources:   msg.Sources,
				CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	duration := time.Since(start)
	log.Info().
		Int("message_count", len(userMessages)).
		Int("total_messages", len(messages)).
		Dur("duration_ms", duration).
		Msg("Get messages request completed")

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
