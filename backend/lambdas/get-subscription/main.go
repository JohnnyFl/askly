package main

import (
	"context"
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

type SubscriptionResponse struct {
	Email                  string     `json:"email"`
	SubscriptionTier       string     `json:"subscription_tier"`
	SubscriptionStatus     string     `json:"subscription_status"`
	QuestionsUsedThisMonth int        `json:"questions_used_this_month"`
	QuestionsLimit         int        `json:"questions_limit"`
	ChannelsLimit          int        `json:"channels_limit"`
	SelectedChannelID      string     `json:"selected_channel_id,omitempty"`
	CurrentPeriodStart     *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd       *time.Time `json:"current_period_end,omitempty"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle OPTIONS preflight request
	if request.HTTPMethod == "OPTIONS" {
		return response.Options()
	}

	// Initialize logger
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	log.Info().Msg("Get subscription request received")

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

	// Initialize user repository
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize user repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	// Get user
	var user models.User
	if err := userRepo.GetUserByID(ctx, userID, &user); err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Check if billing period needs reset
	if user.NeedsReset() {
		log.Info().Msg("Billing period ended, resetting usage")
		now := time.Now()
		nextMonth := now.AddDate(0, 1, 0)
		user.QuestionsUsedThisMonth = 0
		user.CurrentPeriodStart = &now
		user.CurrentPeriodEnd = &nextMonth
		user.UpdatedAt = now

		if err := userRepo.SaveUser(ctx, &user); err != nil {
			log.Error().Err(err).Msg("Failed to reset billing period")
			// Continue anyway, not critical
		}
	}

	log.Info().
		Str("tier", user.SubscriptionTier).
		Int("questions_used", user.QuestionsUsedThisMonth).
		Int("questions_limit", user.QuestionsLimit).
		Msg("Returning subscription data")

	return response.JSON(http.StatusOK, SubscriptionResponse{
		Email:                  user.Email,
		SubscriptionTier:       user.SubscriptionTier,
		SubscriptionStatus:     user.SubscriptionStatus,
		QuestionsUsedThisMonth: user.QuestionsUsedThisMonth,
		QuestionsLimit:         user.QuestionsLimit,
		ChannelsLimit:          user.ChannelsLimit,
		SelectedChannelID:      user.SelectedChannelID,
		CurrentPeriodStart:     user.CurrentPeriodStart,
		CurrentPeriodEnd:       user.CurrentPeriodEnd,
	})
}

func main() {
	lambda.Start(handler)
}
