package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type TranscriptionsResponse struct {
	Transcriptions []TranscriptionResponse `json:"transcriptions"`
}

type TranscriptionResponse struct {
	ID         string `json:"id"`
	VideoTitle string `json:"video_title"`
	VideoURL   string `json:"video_url"`
	UploadDate string `json:"upload_date"`
	Status     string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify JWT token - only admins can list transcriptions
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

	// Only admins can list transcriptions
	if !claims.IsAdmin() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
	}

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	channelPK := "CHANNEL#" + channelID
	log.Printf("INFO: Admin %s listing transcriptions for channel %s", claims.UserID, channelPK)

	// Initialize repository
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize transcription repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Get all transcriptions for this channel
	var transcriptions []*models.Transcription
	if err := transcriptionRepo.ListTranscriptionsByChannel(ctx, channelPK, &transcriptions); err != nil {
		log.Printf("ERROR: Failed to list transcriptions: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch transcriptions"})
	}

	log.Printf("INFO: Found %d transcriptions for channel %s", len(transcriptions), channelPK)

	// Convert to response format
	transcriptionResponses := make([]TranscriptionResponse, len(transcriptions))
	for i, transcription := range transcriptions {
		transcriptionResponses[i] = TranscriptionResponse{
			ID:         strings.TrimPrefix(transcription.PK, "TRANSCRIPTION#"),
			VideoTitle: transcription.VideoTitle,
			VideoURL:   transcription.VideoURL,
			UploadDate: transcription.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
			Status:     transcription.Status,
		}
	}

	return response(http.StatusOK, TranscriptionsResponse{
		Transcriptions: transcriptionResponses,
	})
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
