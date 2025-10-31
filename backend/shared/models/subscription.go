package models

import "time"

// Subscription represents a paid subscription in the system
type Subscription struct {
	PK                   string    `dynamodbav:"PK" json:"id"` // subscription_id
	SK                   string    `dynamodbav:"SK" json:"-"`  // "SUBSCRIPTION"
	UserID               string    `dynamodbav:"user_id" json:"user_id"`
	Tier                 string    `dynamodbav:"tier" json:"tier"`     // "pro", "premium"
	Status               string    `dynamodbav:"status" json:"status"` // "active", "canceled", "past_due", "paused"
	PaddleSubscriptionID string    `dynamodbav:"paddle_subscription_id" json:"paddle_subscription_id"`
	PaddlePriceID        string    `dynamodbav:"paddle_price_id" json:"paddle_price_id"`
	BillingCycle         string    `dynamodbav:"billing_cycle" json:"billing_cycle"` // "monthly", "annual"
	CurrentPeriodStart   time.Time `dynamodbav:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd     time.Time `dynamodbav:"current_period_end" json:"current_period_end"`
	CancelAtPeriodEnd    bool      `dynamodbav:"cancel_at_period_end" json:"cancel_at_period_end"`
	CreatedAt            time.Time `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt            time.Time `dynamodbav:"updated_at" json:"updated_at"`
}

// NewSubscription creates a new Subscription with generated ID
func NewSubscription(userID, tier, paddleSubscriptionID, paddlePriceID string, periodStart, periodEnd time.Time) *Subscription {
	now := time.Now()

	// Determine billing cycle from price ID
	billingCycle := "monthly"
	if isAnnualPriceID(paddlePriceID) {
		billingCycle = "annual"
	}

	return &Subscription{
		PK:                   GenerateID(),
		SK:                   "SUBSCRIPTION",
		UserID:               userID,
		Tier:                 tier,
		Status:               "active",
		PaddleSubscriptionID: paddleSubscriptionID,
		PaddlePriceID:        paddlePriceID,
		BillingCycle:         billingCycle,
		CurrentPeriodStart:   periodStart,
		CurrentPeriodEnd:     periodEnd,
		CancelAtPeriodEnd:    false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// IsActive checks if the subscription is active and not expired
func (s *Subscription) IsActive() bool {
	return s.Status == "active" && time.Now().Before(s.CurrentPeriodEnd)
}

// isAnnualPriceID checks if the price ID corresponds to an annual subscription
// This will be enhanced with actual env var checks in the Paddle service
func isAnnualPriceID(priceID string) bool {
	// TODO: Check against annual price IDs from environment variables
	// For now, we'll check if the price ID contains "annual" or "year"
	return false
}
