package repository

import (
	"context"
	"fmt"
	"os"

	"askly/backend/shared/models"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

// SubscriptionRepository handles subscription data operations
type SubscriptionRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewSubscriptionRepository creates a new subscription repository
// NOTE: Can use either a separate subscriptions table OR the users table
// If DYNAMODB_SUBSCRIPTIONS_TABLE is not set, falls back to askly-subscriptions
func NewSubscriptionRepository(ctx context.Context) (*SubscriptionRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamodb repository: %w", err)
	}

	tableName := os.Getenv("DYNAMODB_SUBSCRIPTIONS_TABLE")
	if tableName == "" {
		// Default to subscriptions table (can be changed to users table if needed)
		tableName = "askly-subscriptions"
	}

	return &SubscriptionRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateSubscription creates a new subscription record
func (r *SubscriptionRepository) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	if err := r.db.PutItem(ctx, r.tableName, sub); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

// GetSubscriptionByID retrieves a subscription by its ID
func (r *SubscriptionRepository) GetSubscriptionByID(ctx context.Context, subID string, result *models.Subscription) error {
	if err := r.db.GetItem(ctx, r.tableName, subID, "SUBSCRIPTION", result); err != nil {
		return fmt.Errorf("failed to get subscription by ID: %w", err)
	}
	return nil
}

// GetSubscriptionByUserID retrieves a subscription by user ID using GSI
func (r *SubscriptionRepository) GetSubscriptionByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	items, err := r.db.QueryByGSIRaw(ctx, r.tableName, "user-index", "user_id", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscription by user ID: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("subscription not found for user: %s", userID)
	}

	var subscription models.Subscription
	if err := attributevalue.UnmarshalMap(items[0], &subscription); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	return &subscription, nil
}

// GetSubscriptionByPaddleID retrieves a subscription by Paddle subscription ID using GSI
func (r *SubscriptionRepository) GetSubscriptionByPaddleID(ctx context.Context, paddleSubID string) (*models.Subscription, error) {
	items, err := r.db.QueryByGSIRaw(ctx, r.tableName, "paddle-subscription-index", "paddle_subscription_id", paddleSubID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscription by paddle ID: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("subscription not found for paddle ID: %s", paddleSubID)
	}

	var subscription models.Subscription
	if err := attributevalue.UnmarshalMap(items[0], &subscription); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	return &subscription, nil
}

// UpdateSubscription updates subscription attributes
func (r *SubscriptionRepository) UpdateSubscription(ctx context.Context, subID string, updates map[string]interface{}) error {
	if err := r.db.UpdateItem(ctx, r.tableName, subID, "SUBSCRIPTION", updates); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}
