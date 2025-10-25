package response

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// CORS headers for all responses
var corsHeaders = map[string]string{
	"Content-Type":                     "application/json",
	"Access-Control-Allow-Origin":      "*",
	"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
	"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Amz-Date, X-Api-Key, X-Amz-Security-Token",
	"Access-Control-Allow-Credentials": "true",
}

// JSON creates a JSON response with proper CORS headers
func JSON(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	jsonBody, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    corsHeaders,
		Body:       string(jsonBody),
	}, nil
}

// Options handles preflight OPTIONS requests
func Options() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    corsHeaders,
		Body:       "",
	}, nil
}

// Error creates an error response with proper CORS headers
func Error(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return JSON(statusCode, map[string]string{
		"error": message,
	})
}
