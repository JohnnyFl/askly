package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"askly/backend/shared/logger"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ChannelResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	YoutubeChannelID string `json:"youtube_channel_id"`
	Description      string `json:"description"`
	ThumbnailURL     string `json:"thumbnail_url"`
	YoutubeURL       string `json:"youtube_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Get channel request received")

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

	log.Debug().Str("email", email).Msg("User requesting channel details")

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Error().Msg("Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	// Initialize repository
	log.Debug().Msg("Initializing channel repository")
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Get channel
	log.Debug().Str("channel_id", channelID).Msg("Fetching channel from database")
	var channel models.Channel
	if err := channelRepo.GetChannelByID(ctx, channelID, &channel); err != nil {
		log.Error().Err(err).Str("channel_id", channelID).Msg("Channel not found")
		return response(http.StatusNotFound, ErrorResponse{Error: "Channel not found"})
	}

	log.Info().Str("channel_name", channel.Name).Msg("Channel retrieved successfully")

	// Convert to response format
	channelResponse := ChannelResponse{
		ID:               channel.PK,
		Name:             channel.Name,
		YoutubeChannelID: channel.YoutubeChannelID,
		Description:      channel.Description,
		ThumbnailURL:     channel.ThumbnailURL,
		YoutubeURL:       channel.YoutubeURL,
	}

	duration := time.Since(start)
	log.Info().
		Str("channel_name", channel.Name).
		Dur("duration_ms", duration).
		Msg("Get channel request completed")

	return response(http.StatusOK, channelResponse)
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
