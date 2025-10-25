package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"askly/backend/shared/logger"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"
	"askly/backend/shared/response"
	"askly/backend/shared/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type UploadRequest struct {
	VideoTitle    string `json:"video_title"`
	VideoID       string `json:"video_id"`
	VideoURL      string `json:"video_url"`
	Transcription string `json:"transcription"`
}

type UploadResponse struct {
	Message         string `json:"message"`
	TranscriptionID string `json:"transcription_id"`
	ChunksCreated   int    `json:"chunks_created"`
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
	log.Info().Msg("Upload transcription request received")

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

	log.Debug().Str("email", email).Msg("Admin uploading transcription")

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Error().Msg("Channel ID missing from path parameters")
		return response.Error(http.StatusBadRequest, "Channel ID is required")
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	// Parse request body
	var req UploadRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response.Error(http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}

	// Validate input
	if req.VideoTitle == "" || req.Transcription == "" {
		log.Warn().Msg("Missing required fields in request")
		return response.Error(http.StatusBadRequest, "video_title and transcription are required")
	}

	log.Info().
		Str("video_title", req.VideoTitle).
		Int("transcription_length", len(req.Transcription)).
		Msg("Starting transcription upload")

	// Initialize repositories
	log.Debug().Msg("Initializing repositories")
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize transcription repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	chunkRepo, err := repository.NewChunkRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize chunk repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	// Verify channel exists
	var channel models.Channel
	log.Debug().Str("channel_id", channelID).Msg("Verifying channel exists")
	if err := channelRepo.GetChannelByID(ctx, channelID, &channel); err != nil {
		log.Error().Err(err).Str("channel_id", channelID).Msg("Channel not found")
		return response.Error(http.StatusNotFound, fmt.Sprintf("Channel not found: %v", err))
	}

	log.Info().Str("channel_name", channel.Name).Msg("Channel verified")

	// Initialize services
	log.Debug().Msg("Initializing AI services")
	chunkingService := services.NewChunkingService()
	bedrockService, err := services.NewBedrockService(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Bedrock service")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize AI service: %v", err))
	}

	// Create transcription record
	transcription := models.NewTranscription(
		channelID,
		req.VideoTitle,
		req.VideoURL,
		req.VideoID,
		req.Transcription,
		adminID,
	)

	log.Debug().Msg("Creating transcription record")
	if err := transcriptionRepo.CreateTranscription(ctx, transcription); err != nil {
		log.Error().Err(err).Msg("Failed to create transcription in DynamoDB")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to create transcription: %v", err))
	}

	transcriptionID := transcription.GetID()
	ctx = logger.WithTranscriptionID(ctx, transcriptionID)
	log = logger.FromContext(ctx)
	log.Info().Str("transcription_id", transcriptionID).Msg("Transcription record created")

	// Chunk the transcription
	chunkStart := time.Now()
	chunks := chunkingService.ChunkText(req.Transcription)
	log.Info().
		Int("chunk_count", len(chunks)).
		Dur("chunking_duration_ms", time.Since(chunkStart)).
		Msg("Text chunked successfully")

	// Process each chunk: generate embedding and store
	successfulChunks := 0
	failedChunks := 0
	for i, chunkText := range chunks {
		chunkLog := log.With().Int("chunk_index", i).Int("total_chunks", len(chunks)).Logger()
		chunkLog.Debug().Int("chunk_length", len(chunkText)).Msg("Processing chunk")

		// Generate embedding for this chunk
		embeddingStart := time.Now()
		embedding, err := bedrockService.GenerateEmbedding(ctx, chunkText)
		if err != nil {
			chunkLog.Error().Err(err).Msg("Failed to generate embedding")
			failedChunks++
			continue
		}

		embeddingDuration := time.Since(embeddingStart)

		// Convert embedding to bytes for storage
		embeddingBytes := services.EmbeddingToBytes(embedding)
		chunkLog.Debug().
			Int("embedding_size", len(embeddingBytes)).
			Int("embedding_dimensions", len(embedding)).
			Dur("embedding_duration_ms", embeddingDuration).
			Msg("Embedding generated")

		// Create chunk record
		chunk := models.NewTranscriptionChunk(
			transcription.PK,
			channelID,
			chunkText,
			i,
			embeddingBytes,
		)

		// Store chunk
		if err := chunkRepo.CreateChunk(ctx, chunk); err != nil {
			chunkLog.Error().Err(err).Msg("Failed to store chunk in DynamoDB")
			failedChunks++
			continue
		}

		successfulChunks++
		chunkLog.Info().Msg("Chunk processed and stored successfully")
	}

	// Log final upload summary
	duration := time.Since(start)
	log.Info().
		Int("successful_chunks", successfulChunks).
		Int("failed_chunks", failedChunks).
		Int("total_chunks", len(chunks)).
		Dur("total_duration_ms", duration).
		Msg("Transcription upload complete")

	return response.JSON(http.StatusCreated, UploadResponse{
		Message:         fmt.Sprintf("Transcription uploaded and processed successfully (%d/%d chunks)", successfulChunks, len(chunks)),
		TranscriptionID: transcriptionID,
		ChunksCreated:   successfulChunks,
	})
}

func main() {
	lambda.Start(handler)
}
