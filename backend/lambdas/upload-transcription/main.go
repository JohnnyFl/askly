package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("ERROR: Failed to extract token: %v", err)
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Missing or invalid authorization header"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		log.Printf("ERROR: Token validation failed: %v", err)
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
	}

	// Only admins can upload transcriptions
	if !claims.IsAdmin() {
		log.Printf("ERROR: Non-admin attempted to upload transcription, role: %s", claims.Role)
		return response(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
	}

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Printf("ERROR: Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Parse request body
	var req UploadRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("ERROR: Failed to parse request body: %v", err)
		return response(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid request body: %v", err)})
	}

	// Validate input
	if req.VideoTitle == "" || req.Transcription == "" {
		log.Printf("ERROR: Missing required fields in request")
		return response(http.StatusBadRequest, ErrorResponse{Error: "video_title and transcription are required"})
	}

	log.Printf("INFO: Admin %s uploading transcription for channel %s, video: %s (%d chars)",
		claims.UserID, channelID, req.VideoTitle, len(req.Transcription))

	// Initialize repositories
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize channel repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize transcription repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	chunkRepo, err := repository.NewChunkRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize chunk repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	// Verify channel exists
	var channel models.Channel
	channelPK := "CHANNEL#" + channelID
	if err := channelRepo.GetChannelByID(ctx, channelPK, &channel); err != nil {
		log.Printf("ERROR: Channel not found: %s, error: %v", channelPK, err)
		return response(http.StatusNotFound, ErrorResponse{Error: fmt.Sprintf("Channel not found: %v", err)})
	}

	log.Printf("INFO: Found channel: %s", channel.Name)

	// Initialize services
	chunkingService := services.NewChunkingService()
	bedrockService, err := services.NewBedrockService(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize Bedrock service: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize AI service: %v", err)})
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
		log.Printf("ERROR: Failed to create transcription in DynamoDB: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create transcription: %v", err)})
	}

	transcriptionID := transcription.GetID()
	log.Printf("INFO: Created transcription record: %s", transcriptionID)

	// Chunk the transcription
	chunks := chunkingService.ChunkText(req.Transcription)
	log.Printf("INFO: Split transcription into %d chunks", len(chunks))

	// Process each chunk: generate embedding and store
	successfulChunks := 0
	for i, chunkText := range chunks {
		log.Printf("INFO: Processing chunk %d/%d (length: %d chars)", i+1, len(chunks), len(chunkText))

		// Generate embedding for this chunk
		embedding, err := bedrockService.GenerateEmbedding(ctx, chunkText)
		if err != nil {
			log.Printf("ERROR: Failed to generate embedding for chunk %d: %v", i, err)
			continue
		}

		// Convert embedding to bytes for storage
		embeddingBytes := services.EmbeddingToBytes(embedding)
		log.Printf("INFO: Generated embedding for chunk %d (size: %d bytes)", i, len(embeddingBytes))

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
			log.Printf("ERROR: Failed to store chunk %d in DynamoDB: %v", i, err)
			continue
		}

		successfulChunks++
		log.Printf("INFO: Successfully stored chunk %d", i)
	}

	log.Printf("INFO: Transcription upload complete. Successfully processed %d/%d chunks", successfulChunks, len(chunks))

	return response(http.StatusCreated, UploadResponse{
		Message:         fmt.Sprintf("Transcription uploaded and processed successfully (%d/%d chunks)", successfulChunks, len(chunks)),
		TranscriptionID: transcriptionID,
		ChunksCreated:   successfulChunks,
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
