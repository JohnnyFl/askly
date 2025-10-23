package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"askly/backend/shared/logger"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type DeleteResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Delete transcription request received")

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

	log.Debug().Str("email", email).Msg("Admin deleting transcription")

	// Get channel ID and transcription ID from path parameters
	channelID := request.PathParameters["id"]
	transcriptionID := request.PathParameters["transcriptionId"]

	if channelID == "" || transcriptionID == "" {
		log.Error().
			Str("channel_id", channelID).
			Str("transcription_id", transcriptionID).
			Msg("Missing required path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID and Transcription ID are required"})
	}

	// Add IDs to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	ctx = logger.WithTranscriptionID(ctx, transcriptionID)
	log = logger.FromContext(ctx)

	log.Info().
		Str("transcription_id", transcriptionID).
		Msg("Admin deleting transcription")

	// Initialize repositories
	log.Debug().Msg("Initializing repositories")
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize transcription repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	chunkRepo, err := repository.NewChunkRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize chunk repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Delete all chunks associated with this transcription
	log.Debug().Msg("Listing chunks for deletion")
	var chunks []*models.TranscriptionChunk
	chunksDeleted := 0
	chunksFailed := 0

	if err := chunkRepo.ListChunksByTranscription(ctx, transcriptionID, &chunks); err != nil {
		log.Warn().Err(err).Str("transcription_id", transcriptionID).Msg("Failed to list chunks for transcription")
	} else {
		log.Info().Int("chunk_count", len(chunks)).Msg("Deleting chunks for transcription")
		for i, chunk := range chunks {
			if err := chunkRepo.DeleteChunk(ctx, chunk.PK); err != nil {
				log.Error().
					Err(err).
					Str("chunk_pk", chunk.PK).
					Int("chunk_index", i).
					Msg("Failed to delete chunk")
				chunksFailed++
			} else {
				chunksDeleted++
				log.Debug().
					Str("chunk_pk", chunk.PK).
					Int("chunk_index", i).
					Msg("Chunk deleted successfully")
			}
		}
	}

	log.Info().
		Int("chunks_deleted", chunksDeleted).
		Int("chunks_failed", chunksFailed).
		Msg("Chunk deletion completed")

	// Delete the transcription itself
	log.Debug().Msg("Deleting transcription record")
	if err := transcriptionRepo.DeleteTranscription(ctx, transcriptionID); err != nil {
		log.Error().Err(err).Str("transcription_id", transcriptionID).Msg("Failed to delete transcription")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete transcription"})
	}

	duration := time.Since(start)
	log.Info().
		Str("transcription_id", transcriptionID).
		Int("chunks_deleted", chunksDeleted).
		Int("chunks_failed", chunksFailed).
		Dur("duration_ms", duration).
		Msg("Transcription deleted successfully")

	return response(http.StatusOK, DeleteResponse{
		Message: "Transcription deleted successfully",
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
