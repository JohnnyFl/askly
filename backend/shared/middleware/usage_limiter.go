package middleware

import (
	"context"
	"fmt"
	"time"

	"askly/backend/shared/models"
	"askly/backend/shared/repository"
)

// TierLimits defines the limits for each subscription tier
type TierLimits struct {
	Tier              string
	QuestionsPerMonth int
	ChannelsMax       int
}

// SubscriptionTiers maps tier names to their limits
var SubscriptionTiers = map[string]TierLimits{
	"free": {
		Tier:              "free",
		QuestionsPerMonth: 5,
		ChannelsMax:       1, // Must select ONE channel
	},
	"pro": {
		Tier:              "pro",
		QuestionsPerMonth: 50,
		ChannelsMax:       2, // Both available channels
	},
	"premium": {
		Tier:              "premium",
		QuestionsPerMonth: 200,
		ChannelsMax:       -1, // Unlimited (all current + future channels)
	},
}

// UsageLimiter handles subscription usage limits and enforcement
type UsageLimiter struct {
	userRepo *repository.UserRepository
}

// NewUsageLimiter creates a new UsageLimiter instance
func NewUsageLimiter(ctx context.Context) (*UsageLimiter, error) {
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user repository: %w", err)
	}

	return &UsageLimiter{
		userRepo: userRepo,
	}, nil
}

// CheckChannelAccess verifies if user can access a specific channel
func (ul *UsageLimiter) CheckChannelAccess(ctx context.Context, userID, channelID string) (bool, error) {
	var user models.User
	if err := ul.userRepo.GetUserByID(ctx, userID, &user); err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return user.CanAccessChannel(channelID), nil
}

// CheckQuestionLimit verifies user can ask another question
// Returns: canAsk, remainingQuestions, error
func (ul *UsageLimiter) CheckQuestionLimit(ctx context.Context, userID string) (bool, int, error) {
	var user models.User
	if err := ul.userRepo.GetUserByID(ctx, userID, &user); err != nil {
		return false, 0, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if we need to reset monthly counter
	if user.NeedsReset() {
		if err := ul.ResetMonthlyUsage(ctx, userID); err != nil {
			return false, 0, fmt.Errorf("failed to reset usage: %w", err)
		}
		// Reload user after reset
		if err := ul.userRepo.GetUserByID(ctx, userID, &user); err != nil {
			return false, 0, fmt.Errorf("failed to reload user: %w", err)
		}
	}

	canAsk := user.CanAskQuestion()
	remaining := user.RemainingQuestions()

	return canAsk, remaining, nil
}

// IncrementQuestionCount increments the question counter for a user
// Note: This method is not currently used. Usage tracking is handled directly in lambdas.
func (ul *UsageLimiter) IncrementQuestionCount(ctx context.Context, userID string) error {
	// Get user
	var user models.User
	if err := ul.userRepo.GetUserByID(ctx, userID, &user); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Increment count
	user.QuestionsUsedThisMonth++
	user.UpdatedAt = time.Now()

	// Update user - Note: This requires SaveUser method in repository
	// For atomic updates, implement in the lambda directly
	return fmt.Errorf("not implemented - use direct lambda implementation for atomic updates")
}

// ResetMonthlyUsage resets the monthly question counter and updates billing period
func (ul *UsageLimiter) ResetMonthlyUsage(ctx context.Context, userID string) error {
	now := time.Now()
	nextPeriodEnd := now.AddDate(0, 1, 0) // Add 1 month

	updates := map[string]interface{}{
		"questions_used_this_month": 0,
		"current_period_start":      now,
		"current_period_end":        nextPeriodEnd,
	}

	if err := ul.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to reset monthly usage: %w", err)
	}

	return nil
}

// SetChannelSelection sets the selected channel for free tier users
func (ul *UsageLimiter) SetChannelSelection(ctx context.Context, userID, channelID string) error {
	var user models.User
	if err := ul.userRepo.GetUserByID(ctx, userID, &user); err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Only allow for free tier users
	if user.SubscriptionTier != "free" {
		return fmt.Errorf("channel selection only available for free tier")
	}

	// Only allow setting once
	if user.SelectedChannelID != "" && user.SelectedChannelID != channelID {
		return fmt.Errorf("channel already selected, upgrade to access both channels")
	}

	updates := map[string]interface{}{
		"selected_channel_id": channelID,
	}

	if err := ul.userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to set channel selection: %w", err)
	}

	return nil
}

// GetTierLimits returns the limits for a given tier
func GetTierLimits(tier string) TierLimits {
	if limits, ok := SubscriptionTiers[tier]; ok {
		return limits
	}
	return SubscriptionTiers["free"] // Default to free
}
