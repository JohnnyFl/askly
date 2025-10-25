package main

import (
	"context"
	"net/http"
	"time"

	"askly/backend/shared/logger"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"
	"askly/backend/shared/response"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ListChannelsResponse struct {
	Channels []ChannelResponse `json:"channels"`
}

type ChannelResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	YoutubeChannelID string `json:"youtube_channel_id"`
	Description      string `json:"description"`
	ThumbnailURL     string `json:"thumbnail_url"`
	AvatarURL        string `json:"avatar_url"`
	YoutubeURL       string `json:"youtube_url"`
	CreatedAt        string `json:"created_at"`
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
	log.Info().Msg("List channels request received")

	// Extract user info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response.Error(http.StatusUnauthorized, "Unauthorized")
	}

	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	// Add user ID to logger context
	ctx = logger.WithUserID(ctx, userID)
	log = logger.FromContext(ctx)

	log.Debug().Str("email", email).Msg("Admin listing channels")

	// Initialize repository
	log.Debug().Msg("Initializing channel repository")
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response.Error(http.StatusInternalServerError, "Internal server error")
	}

	// Get all channels
	log.Debug().Msg("Fetching channels from database")
	var channels []*models.Channel
	if err := channelRepo.ListChannels(ctx, &channels); err != nil {
		log.Error().Err(err).Msg("Failed to fetch channels")
		return response.Error(http.StatusInternalServerError, "Failed to fetch channels")
	}

	log.Info().Int("channel_count", len(channels)).Msg("Channels retrieved successfully")

	// Convert to response format
	log.Debug().Msg("Converting channels to response format")
	channelResponses := make([]ChannelResponse, len(channels))
	for i, channel := range channels {
		channelResponses[i] = ChannelResponse{
			ID:               channel.PK,
			Name:             channel.Name,
			YoutubeChannelID: channel.YoutubeChannelID,
			Description:      channel.Description,
			ThumbnailURL:     channel.ThumbnailURL,
			AvatarURL:        channel.ThumbnailURL,
			YoutubeURL:       channel.YoutubeURL,
			CreatedAt:        channel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	duration := time.Since(start)
	log.Info().
		Int("channel_count", len(channelResponses)).
		Dur("duration_ms", duration).
		Msg("List channels request completed")

	return response.JSON(http.StatusOK, ListChannelsResponse{
		Channels: channelResponses,
	})
}

func main() {
	lambda.Start(handler)
}
