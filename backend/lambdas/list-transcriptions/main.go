package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"askly/backend/shared/logger"
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
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("List transcriptions request received")

	// Extract admin info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	adminID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	// Add admin ID to logger context
	ctx = logger.WithAdminID(ctx, adminID)
	log = logger.FromContext(ctx)

	log.Debug().Str("email", email).Msg("Admin listing transcriptions")

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Error().Msg("Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	channelPK := "CHANNEL#" + channelID
	log.Info().Str("channel_pk", channelPK).Msg("Admin listing transcriptions for channel")

	// Initialize repository
	log.Debug().Msg("Initializing transcription repository")
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize transcription repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Get all transcriptions for this channel
	log.Debug().Msg("Fetching transcriptions from database")
	var transcriptions []*models.Transcription
	if err := transcriptionRepo.ListTranscriptionsByChannel(ctx, channelPK, &transcriptions); err != nil {
		log.Error().Err(err).Str("channel_pk", channelPK).Msg("Failed to list transcriptions")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch transcriptions"})
	}

	log.Info().Int("transcription_count", len(transcriptions)).Msg("Transcriptions retrieved successfully")

	// Convert to response format
	log.Debug().Msg("Converting transcriptions to response format")
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

	duration := time.Since(start)
	log.Info().
		Int("transcription_count", len(transcriptionResponses)).
		Dur("duration_ms", duration).
		Msg("List transcriptions request completed")

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
