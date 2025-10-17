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

type SendMessageRequest struct {
	Question string `json:"question"`
}

type SendMessageResponse struct {
	MessageID string                 `json:"message_id"`
	Question  string                 `json:"question"`
	Answer    string                 `json:"answer"`
	Sources   []models.MessageSource `json:"sources"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Global vector search service (initialized on cold start)
var vectorSearch *services.VectorSearchService

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify JWT token
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

	// Only users can send messages
	if !claims.IsUser() {
		log.Printf("ERROR: Non-user attempted to send message, role: %s", claims.Role)
		return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
	}

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Printf("ERROR: Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Parse request body
	var req SendMessageRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("ERROR: Failed to parse request body: %v", err)
		return response(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid request body: %v", err)})
	}

	// Validate input
	if req.Question == "" {
		log.Printf("ERROR: Empty question in request")
		return response(http.StatusBadRequest, ErrorResponse{Error: "question is required"})
	}

	log.Printf("INFO: Processing question for channel %s, user %s: %.100s", channelID, claims.UserID, req.Question)

	// Initialize repositories
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize channel repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize message repository: %v", err)
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

	// Initialize vector search service if not already done
	if vectorSearch == nil {
		log.Printf("INFO: Initializing vector search service (cold start)")
		vectorSearch, err = services.NewVectorSearchService(ctx)
		if err != nil {
			log.Printf("ERROR: Failed to initialize vector search service: %v", err)
			return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize search service: %v", err)})
		}
	}

	// Load chunks if not initialized (happens on cold start)
	if !vectorSearch.IsInitialized() {
		log.Printf("INFO: Loading knowledge base chunks (cold start)")
		if err := vectorSearch.LoadChunks(ctx); err != nil {
			log.Printf("ERROR: Failed to load chunks from DynamoDB: %v", err)
			return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to load knowledge base: %v", err)})
		}
		log.Printf("INFO: Knowledge base loaded successfully")
	}

	// Search for relevant chunks (top 5)
	log.Printf("INFO: Searching for relevant chunks in channel %s", channelID)
	searchResults, err := vectorSearch.SearchByChannel(ctx, channelID, req.Question, 5)
	if err != nil {
		log.Printf("ERROR: Vector search failed: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to search knowledge base: %v", err)})
	}

	log.Printf("INFO: Found %d relevant chunks", len(searchResults))

	// If no results found, inform user
	if len(searchResults) == 0 {
		log.Printf("WARN: No relevant chunks found for question in channel %s", channelID)
		message := models.NewChatMessage(
			claims.UserID,
			channelPK,
			req.Question,
			"I couldn't find any relevant information in this channel's videos to answer your question. The channel may not have any transcriptions uploaded yet.",
			[]models.MessageSource{},
		)

		if err := messageRepo.CreateMessage(ctx, message); err != nil {
			log.Printf("ERROR: Failed to save empty message: %v", err)
			return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to save message: %v", err)})
		}

		log.Printf("INFO: Returned empty answer for message %s", message.GetID())
		return response(http.StatusOK, SendMessageResponse{
			MessageID: message.GetID(),
			Question:  message.Question,
			Answer:    message.Answer,
			Sources:   message.Sources,
		})
	}

	// Initialize transcription repository to get video details
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize transcription repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	// Extract context chunks and sources
	contextChunks := make([]string, len(searchResults))
	sources := make([]models.MessageSource, 0, len(searchResults))

	for i, result := range searchResults {
		contextChunks[i] = result.Chunk.ChunkText
		log.Printf("INFO: Chunk %d - Similarity: %.3f, TranscriptionID: %s", i, result.Similarity, result.Chunk.TranscriptionID)

		// Get transcription details to populate video info
		var transcription models.Transcription
		if err := transcriptionRepo.GetTranscriptionByID(ctx, result.Chunk.TranscriptionID, &transcription); err != nil {
			log.Printf("WARN: Failed to get transcription details for %s: %v", result.Chunk.TranscriptionID, err)
			continue
		}

		sources = append(sources, models.MessageSource{
			VideoTitle:     transcription.VideoTitle,
			VideoURL:       transcription.VideoURL,
			ChunkIndex:     result.Chunk.ChunkIndex,
			RelevanceScore: float64(result.Similarity),
		})
	}

	log.Printf("INFO: Extracted %d sources from %d chunks", len(sources), len(contextChunks))

	// Initialize Bedrock service
	bedrockService, err := services.NewBedrockService(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize Bedrock service: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize AI service: %v", err)})
	}

	// Generate answer using Claude with RAG
	log.Printf("INFO: Calling Bedrock to generate answer")
	answer, err := bedrockService.GenerateAnswer(ctx, req.Question, contextChunks)
	if err != nil {
		log.Printf("ERROR: Bedrock GenerateAnswer failed: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to generate answer: %v", err)})
	}

	log.Printf("INFO: Generated answer (length: %d chars)", len(answer))

	// Create and save message
	message := models.NewChatMessage(
		claims.UserID,
		channelPK,
		req.Question,
		answer,
		sources,
	)

	if err := messageRepo.CreateMessage(ctx, message); err != nil {
		log.Printf("ERROR: Failed to save message to DynamoDB: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to save message: %v", err)})
	}

	log.Printf("INFO: Successfully created message %s", message.GetID())

	return response(http.StatusOK, SendMessageResponse{
		MessageID: message.GetID(),
		Question:  message.Question,
		Answer:    message.Answer,
		Sources:   message.Sources,
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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
