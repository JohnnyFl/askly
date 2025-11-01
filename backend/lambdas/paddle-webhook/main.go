package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"askly/backend/shared/middleware"
	"askly/backend/shared/repository"
	"askly/backend/shared/services"
)

// PaddleWebhookEvent represents the structure of a Paddle webhook event
type PaddleWebhookEvent struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"`
	Data       json.RawMessage `json:"data"`
}

// SubscriptionData represents subscription data from Paddle webhooks
type SubscriptionData struct {
	ID                   string     `json:"id"`
	CustomerID           string     `json:"customer_id"`
	Status               string     `json:"status"`
	Items                []Item     `json:"items"`
	CurrentBillingPeriod Period     `json:"current_billing_period"`
	CustomData           CustomData `json:"custom_data"`
	ScheduledChange      *struct {
		Action      string `json:"action"`
		EffectiveAt string `json:"effective_at"`
	} `json:"scheduled_change,omitempty"`
}

// Item represents a subscription item
type Item struct {
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

// Period represents a billing period
type Period struct {
	StartsAt string `json:"starts_at"`
	EndsAt   string `json:"ends_at"`
}

// CustomData represents custom data attached to a subscription
type CustomData struct {
	UserID string `json:"user_id"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received Paddle webhook")

	// 1. Verify webhook signature
	signature := request.Headers["paddle-signature"]
	if signature == "" {
		signature = request.Headers["Paddle-Signature"] // Try capitalized version
	}

	if !verifyWebhookSignature(request.Body, signature) {
		fmt.Println("Invalid webhook signature")
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       `{"error": "Invalid signature"}`,
		}, nil
	}

	// 2. Parse event
	var event PaddleWebhookEvent
	if err := json.Unmarshal([]byte(request.Body), &event); err != nil {
		fmt.Printf("Failed to parse webhook: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "Invalid JSON: %s"}`, err.Error()),
		}, nil
	}

	fmt.Printf("Processing event: %s (ID: %s)\n", event.EventType, event.EventID)

	// 3. Handle event based on type
	var err error
	switch event.EventType {
	case "subscription.created":
		err = handleSubscriptionCreated(ctx, event.Data)
	case "subscription.updated":
		err = handleSubscriptionUpdated(ctx, event.Data)
	case "subscription.canceled":
		err = handleSubscriptionCanceled(ctx, event.Data)
	case "subscription.past_due":
		err = handleSubscriptionPastDue(ctx, event.Data)
	case "subscription.paused":
		err = handleSubscriptionPaused(ctx, event.Data)
	case "subscription.resumed":
		err = handleSubscriptionResumed(ctx, event.Data)
	case "transaction.completed":
		err = handleTransactionCompleted(ctx, event.Data)
	case "transaction.payment_failed":
		err = handleTransactionPaymentFailed(ctx, event.Data)
	default:
		fmt.Printf("Unhandled event type: %s\n", event.EventType)
	}

	if err != nil {
		fmt.Printf("Error handling %s: %v\n", event.EventType, err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf(`{"error": "%s"}`, err.Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"status": "ok"}`,
	}, nil
}

func handleSubscriptionCreated(ctx context.Context, data json.RawMessage) error {
	var subData SubscriptionData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	userID := subData.CustomData.UserID
	if userID == "" {
		return fmt.Errorf("missing user_id in custom_data")
	}

	fmt.Printf("Creating subscription for user: %s\n", userID)

	// Determine tier from price ID
	paddleService := services.NewPaddleService()
	tier := paddleService.GetTierFromPriceID(subData.Items[0].PriceID)
	limits := middleware.GetTierLimits(tier)

	// Parse billing period
	periodStart, _ := time.Parse(time.RFC3339, subData.CurrentBillingPeriod.StartsAt)
	periodEnd, _ := time.Parse(time.RFC3339, subData.CurrentBillingPeriod.EndsAt)

	// Update user with subscription details
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user repo: %w", err)
	}

	updates := map[string]interface{}{
		"subscription_tier":         tier,
		"subscription_status":       subData.Status,
		"paddle_customer_id":        subData.CustomerID,
		"paddle_subscription_id":    subData.ID,
		"questions_limit":           limits.QuestionsPerMonth,
		"channels_limit":            limits.ChannelsMax,
		"current_period_start":      periodStart,
		"current_period_end":        periodEnd,
		"questions_used_this_month": 0,  // Reset on new subscription
		"selected_channel_id":       "", // Clear channel selection when upgrading from free
	}

	if err := userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	fmt.Printf("Successfully created subscription for user %s (tier: %s)\n", userID, tier)
	return nil
}

func handleSubscriptionUpdated(ctx context.Context, data json.RawMessage) error {
	var subData SubscriptionData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	userID := subData.CustomData.UserID
	if userID == "" {
		return fmt.Errorf("missing user_id in custom_data")
	}

	fmt.Printf("Updating subscription for user: %s\n", userID)

	// Determine tier from price ID
	paddleService := services.NewPaddleService()
	tier := paddleService.GetTierFromPriceID(subData.Items[0].PriceID)
	limits := middleware.GetTierLimits(tier)

	// Parse billing period
	periodStart, _ := time.Parse(time.RFC3339, subData.CurrentBillingPeriod.StartsAt)
	periodEnd, _ := time.Parse(time.RFC3339, subData.CurrentBillingPeriod.EndsAt)

	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user repo: %w", err)
	}

	updates := map[string]interface{}{
		"subscription_tier":    tier,
		"subscription_status":  subData.Status,
		"questions_limit":      limits.QuestionsPerMonth,
		"channels_limit":       limits.ChannelsMax,
		"current_period_start": periodStart,
		"current_period_end":   periodEnd,
	}

	if err := userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	fmt.Printf("Successfully updated subscription for user %s (tier: %s, status: %s)\n", userID, tier, subData.Status)
	return nil
}

func handleSubscriptionCanceled(ctx context.Context, data json.RawMessage) error {
	var subData SubscriptionData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	userID := subData.CustomData.UserID
	if userID == "" {
		return fmt.Errorf("missing user_id in custom_data")
	}

	fmt.Printf("Canceling subscription for user: %s\n", userID)

	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user repo: %w", err)
	}

	// Mark as canceled but don't downgrade yet (access until period end)
	updates := map[string]interface{}{
		"subscription_status": "canceled",
	}

	if err := userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// TODO: Schedule downgrade to free tier at period_end
	// Can use EventBridge scheduled event or check on next login

	fmt.Printf("Successfully canceled subscription for user %s\n", userID)
	return nil
}

func handleSubscriptionPastDue(ctx context.Context, data json.RawMessage) error {
	var subData SubscriptionData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	userID := subData.CustomData.UserID
	if userID == "" {
		return fmt.Errorf("missing user_id in custom_data")
	}

	fmt.Printf("Subscription past_due for user: %s\n", userID)

	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user repo: %w", err)
	}

	// Suspend access (but don't downgrade completely)
	updates := map[string]interface{}{
		"subscription_status": "past_due",
	}

	if err := userRepo.UpdateUser(ctx, userID, updates); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// TODO: Send email notification about payment failure

	fmt.Printf("Marked user %s as past_due\n", userID)
	return nil
}

func handleSubscriptionPaused(ctx context.Context, data json.RawMessage) error {
	return handleSubscriptionPastDue(ctx, data) // Same logic
}

func handleSubscriptionResumed(ctx context.Context, data json.RawMessage) error {
	var subData SubscriptionData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	userID := subData.CustomData.UserID
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"subscription_status": "active",
	}

	return userRepo.UpdateUser(ctx, userID, updates)
}

func handleTransactionCompleted(ctx context.Context, data json.RawMessage) error {
	fmt.Println("Transaction completed successfully")
	// Could log to analytics, update payment history, etc.
	return nil
}

func handleTransactionPaymentFailed(ctx context.Context, data json.RawMessage) error {
	fmt.Println("Transaction payment failed")
	// Could send notification to user
	return nil
}

// verifyWebhookSignature verifies the Paddle webhook signature
func verifyWebhookSignature(body, signature string) bool {
	secret := os.Getenv("PADDLE_WEBHOOK_SECRET")
	if secret == "" {
		fmt.Println("PADDLE_WEBHOOK_SECRET not set")
		return false
	}

	// Paddle signature format: "ts=timestamp;h1=signature"
	// Parse signature header
	var timestamp, sig string
	parts := splitSignature(signature)
	for k, v := range parts {
		if k == "ts" {
			timestamp = v
		} else if k == "h1" {
			sig = v
		}
	}

	if timestamp == "" || sig == "" {
		fmt.Println("Invalid signature format")
		return false
	}

	// Compute expected signature: HMAC-SHA256(timestamp + ":" + body, secret)
	message := timestamp + ":" + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}

func splitSignature(sig string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(sig, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}
	return result
}

func main() {
	lambda.Start(handler)
}
