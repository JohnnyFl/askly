package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ChannelResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	YoutubeChannelID string `json:"youtube_channel_id"`
	Description      string `json:"description"`
	ThumbnailURL     string `json:"thumbnail_url"`
	YoutubeURL       string `json:"youtube_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify JWT token
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	token, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Missing or invalid authorization header"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
	}

	// Only users can get channel details
	if !claims.IsUser() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
	}

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Initialize repository
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Get channel
	var channel models.Channel
	channelPK := "CHANNEL#" + channelID
	if err := channelRepo.GetChannelByID(ctx, channelPK, &channel); err != nil {
		return response(http.StatusNotFound, ErrorResponse{Error: "Channel not found"})
	}

	// Convert to response format
	channelResponse := ChannelResponse{
		ID:               strings.TrimPrefix(channel.PK, "CHANNEL#"),
		Name:             channel.Name,
		YoutubeChannelID: channel.YoutubeChannelID,
		Description:      channel.Description,
		ThumbnailURL:     channel.ThumbnailURL,
		YoutubeURL:       channel.YoutubeURL,
	}

	return response(http.StatusOK, channelResponse)
}

func response(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	jsonBody, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(jsonBody),
	}, nil
}

func main() {
	lambda.Start(handler)
}
