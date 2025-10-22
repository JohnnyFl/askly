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
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Send message request received")

	// Extract user info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	userID, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)

	// Add user ID to logger context
	ctx = logger.WithUserID(ctx, userID)
	log = logger.FromContext(ctx)

	log.Debug().Str("email", email).Msg("User sending message")

	// Get channel ID from path parameters
	channelID := request.PathParameters["id"]
	if channelID == "" {
		log.Error().Msg("Channel ID missing from path parameters")
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	// Parse request body
	var req SendMessageRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid request body: %v", err)})
	}

	// Validate input
	if req.Question == "" {
		log.Warn().Msg("Empty question in request")
		return response(http.StatusBadRequest, ErrorResponse{Error: "question is required"})
	}

	log.Info().
		Str("question_preview", truncate(req.Question, 100)).
		Int("question_length", len(req.Question)).
		Msg("Processing user question")

	// Initialize repositories
	log.Debug().Msg("Initializing repositories")
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize message repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	// Verify channel exists
	var channel models.Channel
	channelPK := "CHANNEL#" + channelID
	log.Debug().Str("channel_pk", channelPK).Msg("Verifying channel exists")
	if err := channelRepo.GetChannelByID(ctx, channelPK, &channel); err != nil {
		log.Error().Err(err).Str("channel_pk", channelPK).Msg("Channel not found")
		return response(http.StatusNotFound, ErrorResponse{Error: fmt.Sprintf("Channel not found: %v", err)})
	}

	log.Info().Str("channel_name", channel.Name).Msg("Channel verified")

	// Initialize vector search service if not already done
	if vectorSearch == nil {
		log.Info().Msg("Initializing vector search service (cold start)")
		vectorSearch, err = services.NewVectorSearchService(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize vector search service")
			return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize search service: %v", err)})
		}
	}

	// Load chunks if not initialized (happens on cold start)
	if !vectorSearch.IsInitialized() {
		loadStart := time.Now()
		log.Info().Msg("Loading knowledge base chunks (cold start)")
		if err := vectorSearch.LoadChunks(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to load chunks from DynamoDB")
			return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to load knowledge base: %v", err)})
		}
		log.Info().Dur("load_duration_ms", time.Since(loadStart)).Msg("Knowledge base loaded successfully")
	}

	// Search for relevant chunks (top 5)
	searchStart := time.Now()
	log.Debug().Msg("Searching for relevant chunks")
	searchResults, err := vectorSearch.SearchByChannel(ctx, channelID, req.Question, 5)
	if err != nil {
		log.Error().Err(err).Msg("Vector search failed")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to search knowledge base: %v", err)})
	}

	log.Info().
		Int("result_count", len(searchResults)).
		Dur("search_duration_ms", time.Since(searchStart)).
		Msg("Vector search completed")

	// If no results found, inform user
	if len(searchResults) == 0 {
		log.Warn().Msg("No relevant chunks found for question")
		message := models.NewChatMessage(
			userID,
			channelPK,
			req.Question,
			"I couldn't find any relevant information in this channel's videos to answer your question. The channel may not have any transcriptions uploaded yet.",
			[]models.MessageSource{},
		)

		if err := messageRepo.CreateMessage(ctx, message); err != nil {
			log.Error().Err(err).Msg("Failed to save empty message")
			return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to save message: %v", err)})
		}

		log.Info().Str("message_id", message.GetID()).Msg("Returned empty answer")
		return response(http.StatusOK, SendMessageResponse{
			MessageID: message.GetID(),
			Question:  message.Question,
			Answer:    message.Answer,
			Sources:   message.Sources,
		})
	}

	// Initialize transcription repository to get video details
	log.Debug().Msg("Initializing transcription repository")
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize transcription repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize repository: %v", err)})
	}

	// Extract context chunks and sources
	log.Debug().Msg("Extracting context chunks and sources")
	contextChunks := make([]string, len(searchResults))
	sources := make([]models.MessageSource, 0, len(searchResults))

	for i, result := range searchResults {
		contextChunks[i] = result.Chunk.ChunkText
		log.Debug().
			Int("chunk_index", i).
			Float32("similarity", result.Similarity).
			Str("transcription_id", result.Chunk.TranscriptionID).
			Msg("Processing search result chunk")

		// Get transcription details to populate video info
		var transcription models.Transcription
		if err := transcriptionRepo.GetTranscriptionByID(ctx, result.Chunk.TranscriptionID, &transcription); err != nil {
			log.Warn().
				Err(err).
				Str("transcription_id", result.Chunk.TranscriptionID).
				Msg("Failed to get transcription details")
			continue
		}

		sources = append(sources, models.MessageSource{
			VideoTitle:     transcription.VideoTitle,
			VideoURL:       transcription.VideoURL,
			ChunkIndex:     result.Chunk.ChunkIndex,
			RelevanceScore: float64(result.Similarity),
		})
	}

	log.Info().
		Int("source_count", len(sources)).
		Int("chunk_count", len(contextChunks)).
		Msg("Extracted sources from chunks")

	// Initialize Bedrock service
	log.Debug().Msg("Initializing Bedrock service")
	bedrockService, err := services.NewBedrockService(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Bedrock service")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to initialize AI service: %v", err)})
	}

	// Generate answer using Claude with RAG
	generateStart := time.Now()
	log.Info().Int("context_chunk_count", len(contextChunks)).Msg("Calling Bedrock to generate answer")
	answer, err := bedrockService.GenerateAnswer(ctx, req.Question, contextChunks)
	if err != nil {
		log.Error().Err(err).Msg("Bedrock GenerateAnswer failed")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to generate answer: %v", err)})
	}

	log.Info().
		Int("answer_length", len(answer)).
		Dur("generate_duration_ms", time.Since(generateStart)).
		Msg("Answer generated successfully")

	// Create and save message
	message := models.NewChatMessage(
		userID,
		channelPK,
		req.Question,
		answer,
		sources,
	)

	log.Debug().Msg("Saving message to DynamoDB")
	if err := messageRepo.CreateMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Failed to save message to DynamoDB")
		return response(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to save message: %v", err)})
	}

	duration := time.Since(start)
	log.Info().
		Str("message_id", message.GetID()).
		Int("source_count", len(sources)).
		Int("answer_length", len(answer)).
		Dur("total_duration_ms", duration).
		Msg("Send message request completed successfully")

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
