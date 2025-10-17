package main

import (
	"context"
	"encoding/json"
	"net/http"

	"askly/backend/shared/auth"
	"askly/backend/shared/models"
	"askly/backend/shared/repository"
	"askly/backend/shared/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type SendMessageRequest struct {
	ChannelID string `json:"channel_id"`
	Question  string `json:"question"`
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
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Missing or invalid authorization header"})
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired token"})
	}

	// Only users can send messages
	if !claims.IsUser() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
	}

	// Parse request body
	var req SendMessageRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	// Validate input
	if req.ChannelID == "" || req.Question == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "channel_id and question are required"})
	}

	// Initialize repositories
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Verify channel exists
	var channel models.Channel
	channelPK := "CHANNEL#" + req.ChannelID
	if err := channelRepo.GetChannelByID(ctx, channelPK, &channel); err != nil {
		return response(http.StatusNotFound, ErrorResponse{Error: "Channel not found"})
	}

	// Initialize vector search service if not already done
	if vectorSearch == nil {
		vectorSearch, err = services.NewVectorSearchService(ctx)
		if err != nil {
			return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to initialize search service"})
		}
	}

	// Load chunks if not initialized (happens on cold start)
	if !vectorSearch.IsInitialized() {
		if err := vectorSearch.LoadChunks(ctx); err != nil {
			return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to load knowledge base"})
		}
	}

	// Search for relevant chunks (top 5)
	searchResults, err := vectorSearch.SearchByChannel(ctx, req.ChannelID, req.Question, 5)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to search knowledge base"})
	}

	// If no results found, inform user
	if len(searchResults) == 0 {
		message := models.NewChatMessage(
			claims.UserID,
			req.ChannelID,
			req.Question,
			"I couldn't find any relevant information in this channel's videos to answer your question. The channel may not have any transcriptions uploaded yet.",
			[]models.MessageSource{},
		)

		if err := messageRepo.CreateMessage(ctx, message); err != nil {
			return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save message"})
		}

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
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Extract context chunks and sources
	contextChunks := make([]string, len(searchResults))
	sources := make([]models.MessageSource, 0, len(searchResults))

	for i, result := range searchResults {
		contextChunks[i] = result.Chunk.ChunkText

		// Get transcription details to populate video info
		var transcription models.Transcription
		if err := transcriptionRepo.GetTranscriptionByID(ctx, result.Chunk.TranscriptionID, &transcription); err != nil {
			// If we can't get transcription, skip this source
			continue
		}

		sources = append(sources, models.MessageSource{
			VideoTitle:     transcription.VideoTitle,
			VideoURL:       transcription.VideoURL,
			ChunkIndex:     result.Chunk.ChunkIndex,
			RelevanceScore: float64(result.Similarity),
		})
	}

	// Initialize Bedrock service
	bedrockService, err := services.NewBedrockService(ctx)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to initialize AI service"})
	}

	// Generate answer using Claude with RAG
	answer, err := bedrockService.GenerateAnswer(ctx, req.Question, contextChunks)
	if err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate answer"})
	}

	// Create and save message
	message := models.NewChatMessage(
		claims.UserID,
		req.ChannelID,
		req.Question,
		answer,
		sources,
	)

	if err := messageRepo.CreateMessage(ctx, message); err != nil {
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save message"})
	}

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
