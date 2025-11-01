package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// PaddleService handles all interactions with the Paddle API
type PaddleService struct {
	apiKey      string
	environment string
	baseURL     string
	client      *http.Client
}

// NewPaddleService creates a new Paddle service client
func NewPaddleService() *PaddleService {
	environment := os.Getenv("PADDLE_ENVIRONMENT")
	if environment == "" {
		environment = "sandbox"
	}

	baseURL := "https://api.paddle.com"
	if environment == "sandbox" {
		baseURL = "https://sandbox-api.paddle.com"
	}

	apiKey := os.Getenv("PADDLE_API_KEY")

	return &PaddleService{
		apiKey:      apiKey,
		environment: environment,
		baseURL:     baseURL,
		client:      &http.Client{},
	}
}

// Checkout structures
type CreateCheckoutRequest struct {
	Items         []CheckoutItem    `json:"items"`
	CustomerEmail string            `json:"customer_email,omitempty"`
	CustomData    map[string]string `json:"custom_data,omitempty"`
	SuccessURL    string            `json:"success_url,omitempty"`
}

type CheckoutItem struct {
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

type CreateCheckoutResponse struct {
	Data struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	} `json:"data"`
}

// CreateCheckout creates a new Paddle checkout session
func (ps *PaddleService) CreateCheckout(ctx context.Context, priceID, userID, email string) (string, error) {
	reqBody := CreateCheckoutRequest{
		Items: []CheckoutItem{
			{
				PriceID:  priceID,
				Quantity: 1,
			},
		},
		CustomerEmail: email,
		CustomData: map[string]string{
			"user_id": userID,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ps.baseURL+"/checkouts", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ps.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ps.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("paddle API error (%d): %s", resp.StatusCode, string(body))
	}

	var checkoutResp CreateCheckoutResponse
	if err := json.Unmarshal(body, &checkoutResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return checkoutResp.Data.URL, nil
}

// Subscription structures
type SubscriptionResponse struct {
	Data struct {
		ID                   string `json:"id"`
		Status               string `json:"status"`
		CustomerID           string `json:"customer_id"`
		Items                []Item `json:"items"`
		CurrentBillingPeriod Period `json:"current_billing_period"`
		ScheduledChange      *struct {
			Action      string `json:"action"`
			EffectiveAt string `json:"effective_at"`
		} `json:"scheduled_change,omitempty"`
	} `json:"data"`
}

type Item struct {
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

type Period struct {
	StartsAt string `json:"starts_at"`
	EndsAt   string `json:"ends_at"`
}

// GetSubscription retrieves subscription details from Paddle
func (ps *PaddleService) GetSubscription(ctx context.Context, subscriptionID string) (*SubscriptionResponse, error) {
	url := fmt.Sprintf("%s/subscriptions/%s", ps.baseURL, subscriptionID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ps.apiKey)

	resp, err := ps.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paddle API error (%d): %s", resp.StatusCode, string(body))
	}

	var subResp SubscriptionResponse
	if err := json.Unmarshal(body, &subResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &subResp, nil
}

// CancelSubscription cancels a subscription at the end of the billing period
func (ps *PaddleService) CancelSubscription(ctx context.Context, subscriptionID string) error {
	url := fmt.Sprintf("%s/subscriptions/%s", ps.baseURL, subscriptionID)

	reqBody := map[string]interface{}{
		"scheduled_change": map[string]string{
			"action": "cancel",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+ps.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ps.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("paddle API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTierFromPriceID determines the subscription tier from a Paddle price ID
func (ps *PaddleService) GetTierFromPriceID(priceID string) string {
	proMonthly := os.Getenv("PADDLE_PRO_MONTHLY_PRICE_ID")
	proAnnual := os.Getenv("PADDLE_PRO_ANNUAL_PRICE_ID")
	premiumMonthly := os.Getenv("PADDLE_PREMIUM_MONTHLY_PRICE_ID")
	premiumAnnual := os.Getenv("PADDLE_PREMIUM_ANNUAL_PRICE_ID")

	if priceID == proMonthly || priceID == proAnnual {
		return "pro"
	}
	if priceID == premiumMonthly || priceID == premiumAnnual {
		return "premium"
	}
	return "free"
}
