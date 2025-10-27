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

type TranscriptionResponse struct {
	ID           string `json:"id"`
	FileName     string `json:"file_name"`
	FileSize     int    `json:"file_size"`
	VideoTitle   string `json:"video_title"`
	VideoURL     string `json:"video_url"`
	UploadDate   string `json:"upload_date"`
	Status       string `json:"status"`
	FullText     string `json:"full_text"`
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
	log.Info().Msg("Get transcription request received")

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

	log.Debug().Str("email", email).Msg("Admin fetching transcription")

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Error().Msg("Channel ID missing from path parameters")
		return response.Error(http.StatusBadRequest, "Channel ID is required")
	}

	// Get transcription ID from path parameters
	transcriptionID := request.PathParameters["transcriptionId"]
	if transcriptionID == "" {
		log.Error().Msg("Transcription ID missing from path parameters")
		return response.Error(http.StatusBadRequest, "Transcription ID is required")
	}

	// Add IDs to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	ctx = logger.WithTranscriptionID(ctx, transcriptionID)
	log = logger.FromContext(ctx)

	log.Info().
		Str("channel_id", channelID).
		Str("transcription_id", transcriptionID).
		Msg("Admin fetching single transcription")

	// Initialize repository
	log.Debug().Msg("Initializing transcription repository")
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize transcription repository")
		return response.Error(http.StatusInternalServerError, "Internal server error")
	}

	// Get the transcription
	log.Debug().Msg("Fetching transcription from database")
	var transcription models.Transcription
	if err := transcriptionRepo.GetTranscriptionByID(ctx, transcriptionID, &transcription); err != nil {
		log.Error().Err(err).Str("transcription_id", transcriptionID).Msg("Transcription not found")
		return response.Error(http.StatusNotFound, "Transcription not found")
	}

	// Verify the transcription belongs to the requested channel
	if transcription.ChannelID != channelID {
		log.Warn().
			Str("expected_channel_id", channelID).
			Str("actual_channel_id", transcription.ChannelID).
			Msg("Transcription does not belong to requested channel")
		return response.Error(http.StatusNotFound, "Transcription not found")
	}

	// Use video title as file name, calculate file size from full text
	fileName := transcription.VideoTitle
	if fileName == "" {
		fileName = "Untitled"
	}
	fileSize := len(transcription.FullText)

	// Convert to response format with full text
	transcriptionResponse := TranscriptionResponse{
		ID:         transcription.PK,
		FileName:   fileName,
		FileSize:   fileSize,
		VideoTitle: transcription.VideoTitle,
		VideoURL:   transcription.VideoURL,
		UploadDate: transcription.UploadedAt.Format("2006-01-02T15:04:05Z07:00"),
		Status:     transcription.Status,
		FullText:   transcription.FullText,
	}

	duration := time.Since(start)
	log.Info().
		Str("transcription_id", transcriptionID).
		Int("full_text_length", len(transcription.FullText)).
		Dur("duration_ms", duration).
		Msg("Get transcription request completed")

	return response.JSON(http.StatusOK, transcriptionResponse)
}

func main() {
	lambda.Start(handler)
}
