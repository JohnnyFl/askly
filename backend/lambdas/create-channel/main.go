package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"askly/backend/shared/logger"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"
	"askly/backend/shared/response"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type CreateChannelRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	ThumbnailURL     string `json:"thumbnail_url"`
	YoutubeURL       string `json:"youtube_url"`
	YoutubeChannelID string `json:"youtube_channel_id"`
}

type CreateChannelResponse struct {
	Message string         `json:"message"`
	Channel models.Channel `json:"channel"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle OPTIONS preflight request
	if request.HTTPMethod == "OPTIONS" {
		return response.Options()
	}

	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Create channel request received")

	// Extract admin info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response.Error(http.StatusUnauthorized, "Unauthorized")
	}

	adminID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	// Add admin ID to logger context
	ctx = logger.WithAdminID(ctx, adminID)
	log = logger.FromContext(ctx)

	log.Debug().Str("email", email).Msg("Admin creating channel")

	// Parse request body
	var req CreateChannelRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response.Error(http.StatusBadRequest, "Invalid request body")
	}

	// Validate required fields
	if req.Name == "" {
		log.Warn().Msg("Channel name is required")
		return response.Error(http.StatusBadRequest, "Channel name is required")
	}
	if req.YoutubeChannelID == "" {
		log.Warn().Msg("YouTube channel ID is required")
		return response.Error(http.StatusBadRequest, "YouTube channel ID is required")
	}

	// Trim whitespace
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.ThumbnailURL = strings.TrimSpace(req.ThumbnailURL)
	req.YoutubeURL = strings.TrimSpace(req.YoutubeURL)
	req.YoutubeChannelID = strings.TrimSpace(req.YoutubeChannelID)

	log.Info().
		Str("name", req.Name).
		Str("youtube_channel_id", req.YoutubeChannelID).
		Msg("Creating new channel")

	// Initialize repository
	log.Debug().Msg("Initializing channel repository")
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response.Error(http.StatusInternalServerError, "Internal server error")
	}

	// Check if channel with this YouTube ID already exists
	log.Debug().Str("youtube_channel_id", req.YoutubeChannelID).Msg("Checking for existing channel")
	var existingChannel models.Channel
	if err := channelRepo.GetChannelByYoutubeID(ctx, req.YoutubeChannelID, &existingChannel); err != nil {
		log.Error().Err(err).Msg("Failed to check for existing channel")
		return response.Error(http.StatusInternalServerError, "Failed to check for existing channel")
	}

	if existingChannel.PK != "" {
		log.Warn().
			Str("existing_channel_id", existingChannel.PK).
			Str("youtube_channel_id", req.YoutubeChannelID).
			Msg("Channel with this YouTube ID already exists")
		return response.Error(http.StatusConflict, "Channel with this YouTube channel ID already exists")
	}

	// Create new channel
	channel := models.NewChannel(
		req.Name,
		req.Description,
		req.ThumbnailURL,
		req.YoutubeURL,
		req.YoutubeChannelID,
	)

	log.Debug().Str("channel_id", channel.GetID()).Msg("Saving channel to database")
	if err := channelRepo.CreateChannel(ctx, channel); err != nil {
		log.Error().Err(err).Msg("Failed to create channel")
		return response.Error(http.StatusInternalServerError, "Failed to create channel")
	}

	duration := time.Since(start)
	log.Info().
		Str("channel_id", channel.GetID()).
		Str("name", channel.Name).
		Dur("duration_ms", duration).
		Msg("Channel created successfully")

	return response.JSON(http.StatusCreated, CreateChannelResponse{
		Message: "Channel created successfully",
		Channel: *channel,
	})
}

func main() {
	lambda.Start(handler)
}
