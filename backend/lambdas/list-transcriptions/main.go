package main

import (
	"context"
	"net/http"
	"time"

	"askly/backend/shared/logger"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"
	"askly/backend/shared/response"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type TranscriptionsResponse struct {
	Transcriptions []TranscriptionResponse `json:"transcriptions"`
}

type TranscriptionResponse struct {
	ID           string `json:"id"`
	FileName     string `json:"file_name"`
	FileSize     int    `json:"file_size"`
	VideoTitle   string `json:"video_title"`
	VideoURL     string `json:"video_url"`
	UploadDate   string `json:"upload_date"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle OPTIONS preflight request
	if request.HTTPMethod == "OPTIONS" {
		return response.Options()
	}

	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("List transcriptions request received")

	// Extract admin info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response.Error(http.StatusUnauthorized, "Unauthorized")
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
		return response.Error(http.StatusBadRequest, "Channel ID is required")
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	log.Info().Str("channel_id", channelID).Msg("Admin listing transcriptions for channel")

	// Initialize repository
	log.Debug().Msg("Initializing transcription repository")
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize transcription repository")
		return response.Error(http.StatusInternalServerError, "Internal server error")
	}

	// Get all transcriptions for this channel
	log.Debug().Msg("Fetching transcriptions from database")
	var transcriptions []*models.Transcription
	if err := transcriptionRepo.ListTranscriptionsByChannel(ctx, channelID, &transcriptions); err != nil {
		log.Error().Err(err).Str("channel_id", channelID).Msg("Failed to list transcriptions")
		return response.Error(http.StatusInternalServerError, "Failed to fetch transcriptions")
	}

	log.Info().Int("transcription_count", len(transcriptions)).Msg("Transcriptions retrieved successfully")

	// Convert to response format
	log.Debug().Msg("Converting transcriptions to response format")
	transcriptionResponses := make([]TranscriptionResponse, len(transcriptions))
	for i, transcription := range transcriptions {
		// Use video title as file name, calculate approximate file size from full text
		fileName := transcription.VideoTitle
		if fileName == "" {
			fileName = "Untitled"
		}
		fileSize := len(transcription.FullText)

		transcriptionResponses[i] = TranscriptionResponse{
			ID:         transcription.PK,
			FileName:   fileName,
			FileSize:   fileSize,
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

	return response.JSON(http.StatusOK, TranscriptionsResponse{
		Transcriptions: transcriptionResponses,
	})
}

func main() {
	lambda.Start(handler)
}
