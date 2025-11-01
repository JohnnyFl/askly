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
	"askly/backend/shared/response"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type SetChannelRequest struct {
	ChannelID string `json:"channel_id"`
}

type SetChannelResponse struct {
	Message           string `json:"message"`
	SelectedChannelID string `json:"selected_channel_id"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle OPTIONS preflight request
	if request.HTTPMethod == "OPTIONS" {
		return response.Options()
	}

	// Initialize logger
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	log.Info().Msg("Set channel selection request received")

	// Extract user info from Cognito authorizer claims
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response.Error(http.StatusUnauthorized, "Unauthorized")
	}

	userID, _ := claims["sub"].(string)

	// Add user ID to logger context
	ctx = logger.WithUserID(ctx, userID)
	log = logger.FromContext(ctx)

	// Parse request body
	var req SetChannelRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response.Error(http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}

	if req.ChannelID == "" {
		log.Warn().Msg("Empty channel_id in request")
		return response.Error(http.StatusBadRequest, "channel_id is required")
	}

	// Initialize repositories
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize user repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	// Verify channel exists
	var channel models.Channel
	if err := channelRepo.GetChannelByID(ctx, req.ChannelID, &channel); err != nil {
		log.Error().Err(err).Str("channel_id", req.ChannelID).Msg("Channel not found")
		return response.Error(http.StatusNotFound, "Channel not found")
	}

	// Get user
	var user models.User
	if err := userRepo.GetUserByID(ctx, userID, &user); err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Check if user is on free tier
	if user.SubscriptionTier != "free" {
		log.Warn().Str("tier", user.SubscriptionTier).Msg("Non-free user attempting to set channel selection")
		return response.Error(http.StatusBadRequest, "Only free tier users need to select a channel")
	}

	// Update user's selected channel
	user.SelectedChannelID = req.ChannelID
	user.UpdatedAt = time.Now()

	if err := userRepo.SaveUser(ctx, &user); err != nil {
		log.Error().Err(err).Msg("Failed to update user")
		return response.Error(http.StatusInternalServerError, "Failed to save channel selection")
	}

	log.Info().
		Str("channel_id", req.ChannelID).
		Str("channel_name", channel.Name).
		Msg("Channel selection updated successfully")

	return response.JSON(http.StatusOK, SetChannelResponse{
		Message:           "Channel selection saved successfully",
		SelectedChannelID: req.ChannelID,
	})
}

func main() {
	lambda.Start(handler)
}
