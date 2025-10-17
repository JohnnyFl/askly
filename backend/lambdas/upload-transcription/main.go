package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"
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
	// Verify JWT token - only admins can upload
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

	// Only admins can upload transcriptions
	if !claims.IsAdmin() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
	}

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Parse request body
	var req UploadRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate input
	if req.VideoTitle == "" || req.Transcription == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "video_title and transcription are required"})
	}

	// Initialize repositories
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	chunkRepo, err := repository.NewChunkRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Verify channel exists
	var channel models.Channel
	channelPK := "CHANNEL#" + channelID
	if err := channelRepo.GetChannelByID(ctx, channelPK, &channel); err != nil {
		return response(http.StatusNotFound, ErrorResponse{Error: "Channel not found"})
	}

	// Initialize services
	chunkingService := services.NewChunkingService()
	bedrockService, err := services.NewBedrockService(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to initialize AI service"})
	}

	// Create transcription record
	adminID := claims.UserID
	transcription := models.NewTranscription(
		channelPK,
		req.VideoTitle,
		req.VideoURL,
		req.VideoID,
		req.Transcription,
		"ADMIN#"+adminID,
	)

	if err := transcriptionRepo.CreateTranscription(ctx, transcription); err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create transcription"})
	}

	transcriptionID := transcription.GetID()

	// Chunk the transcription
	chunks := chunkingService.ChunkText(req.Transcription)

	// Process each chunk: generate embedding and store
	for i, chunkText := range chunks {
		// Generate embedding for this chunk
		embedding, err := bedrockService.GenerateEmbedding(ctx, chunkText)
		if err != nil {
			// Log error but continue with other chunks
			fmt.Printf("Failed to generate embedding for chunk %d: %v\n", i, err)
			continue
		}

		// Convert embedding to bytes for storage
		embeddingBytes := services.EmbeddingToBytes(embedding)

		// Create chunk record
		chunk := models.NewTranscriptionChunk(
			transcription.PK,
			channelPK,
			chunkText,
			i,
			embeddingBytes,
		)

		// Store chunk
		if err := chunkRepo.CreateChunk(ctx, chunk); err != nil {
			fmt.Printf("Failed to store chunk %d: %v\n", i, err)
			continue
		}
	}

	return response(http.StatusCreated, UploadResponse{
		Message:         "Transcription uploaded and processed successfully",
		TranscriptionID: transcriptionID,
		ChunksCreated:   len(chunks),
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
