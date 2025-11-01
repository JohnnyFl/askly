package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"askly/backend/shared/services"
)

// CheckoutRequest represents the request body for creating a checkout
type CheckoutRequest struct {
	PriceID string `json:"price_id"`
	Email   string `json:"email,omitempty"`
}

// CheckoutResponse represents the response containing the checkout URL
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get user ID from JWT authorizer (Cognito)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "Unauthorized - no claims"}`,
		}, nil
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "Unauthorized - no user ID"}`,
		}, nil
	}

	// Parse request body
	var req CheckoutRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "Invalid request: %s"}`, err.Error()),
		}, nil
	}

	if req.PriceID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error": "price_id is required"}`,
		}, nil
	}

	fmt.Printf("Creating checkout for user %s with price %s\n", userID, req.PriceID)

	// Create Paddle checkout
	paddleService := services.NewPaddleService()
	checkoutURL, err := paddleService.CreateCheckout(ctx, req.PriceID, userID, req.Email)
	if err != nil {
		fmt.Printf("Failed to create checkout: %v\n", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: fmt.Sprintf(`{"error": "Failed to create checkout: %s"}`, err.Error()),
		}, nil
	}

	// Return checkout URL
	response := CheckoutResponse{
		CheckoutURL: checkoutURL,
	}

	responseJSON, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(responseJSON),
	}, nil
}

func main() {
	lambda.Start(handler)
}
