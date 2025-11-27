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

type SendMessageRequest struct {
	Question    string `json:"question"`
	PromptStyle string `json:"prompt_style,omitempty"` // optional: educational, direct, mentor, excited, gojo
}

type SendMessageResponse struct {
	MessageID   string                 `json:"message_id"`
	Question    string                 `json:"question"`
	Answer      string                 `json:"answer"`
	Sources     []models.MessageSource `json:"sources"`
	PromptStyle string                 `json:"prompt_style"` // returns the style used
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Global vector search service (initialized on cold start)
var vectorSearch *services.VectorSearchService

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Handle OPTIONS preflight request
	if request.HTTPMethod == "OPTIONS" {
		return response.Options()
	}

	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Send message request received")

	// Extract user info from Cognito authorizer claims (already validated by API Gateway)
	claims, ok := request.RequestContext.Authorizer["claims"].(map[string]interface{})
	if !ok || claims == nil {
		log.Error().Msg("No Cognito claims found in request context")
		return response.Error(http.StatusUnauthorized, "Unauthorized")
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
		return response.Error(http.StatusBadRequest, "Channel ID is required")
	}

	// Add channel ID to logger context
	ctx = logger.WithChannelID(ctx, channelID)
	log = logger.FromContext(ctx)

	// Parse request body
	var req SendMessageRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		return response.Error(http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
	}

	// Validate input
	if req.Question == "" {
		log.Warn().Msg("Empty question in request")
		return response.Error(http.StatusBadRequest, "question is required")
	}

	// Determine prompt style (default to "direct")
	promptStyle := services.PromptStyle(req.PromptStyle)
	if req.PromptStyle == "" {
		promptStyle = services.PromptStyleDirect
	}
	// Validate prompt style
	validStyles := map[services.PromptStyle]bool{
		services.PromptStyleEducational: true,
		services.PromptStyleDirect:      true,
		services.PromptStyleMentor:      true,
		services.PromptStyleExcited:     true,
		services.PromptStyleGojo:        true,
	}
	if !validStyles[promptStyle] {
		log.Warn().Str("invalid_style", req.PromptStyle).Msg("Invalid prompt style, using default")
		promptStyle = services.PromptStyleDirect
	}

	log.Info().
		Str("question_preview", truncate(req.Question, 100)).
		Int("question_length", len(req.Question)).
		Str("prompt_style", string(promptStyle)).
		Msg("Processing user question")

	// Initialize repositories
	log.Debug().Msg("Initializing repositories")
	userRepo, err := repository.NewUserRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize user repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	messageRepo, err := repository.NewMessageRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize message repository")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
	}

	// Get user to check subscription limits
	var user models.User
	if err := userRepo.GetUserByID(ctx, userID, &user); err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return response.Error(http.StatusNotFound, "User not found")
	}

	// Reset usage if billing period ended
	if user.NeedsReset() {
		log.Info().Msg("Billing period ended, resetting usage")
		now := time.Now()
		nextMonth := now.AddDate(0, 1, 0)
		user.QuestionsUsedThisMonth = 0
		user.CurrentPeriodStart = &now
		user.CurrentPeriodEnd = &nextMonth
		user.UpdatedAt = now

		if err := userRepo.SaveUser(ctx, &user); err != nil {
			log.Error().Err(err).Msg("Failed to reset billing period")
			// Continue anyway
		}
	}

	// Check if user has questions remaining
	if !user.CanAskQuestion() {
		log.Warn().
			Int("used", user.QuestionsUsedThisMonth).
			Int("limit", user.QuestionsLimit).
			Msg("User exceeded question limit")
		return response.Error(http.StatusTooManyRequests, fmt.Sprintf("You've used all %d questions this month. Upgrade to ask more!", user.QuestionsLimit))
	}

	// Check if user can access this channel
	if !user.CanAccessChannel(channelID) {
		log.Warn().
			Str("user_tier", user.SubscriptionTier).
			Str("selected_channel", user.SelectedChannelID).
			Str("requested_channel", channelID).
			Msg("User cannot access this channel")
		return response.Error(http.StatusForbidden, "You don't have access to this channel. Free users can only access one selected channel. Upgrade to Pro for more!")
	}

	// Verify channel exists
	var channel models.Channel
	log.Debug().Str("channel_id", channelID).Msg("Verifying channel exists")
	if err := channelRepo.GetChannelByID(ctx, channelID, &channel); err != nil {
		log.Error().Err(err).Str("channel_id", channelID).Msg("Channel not found")
		return response.Error(http.StatusNotFound, fmt.Sprintf("Channel not found: %v", err))
	}

	log.Info().Str("channel_name", channel.Name).Msg("Channel verified")

	// Initialize vector search service if not already done
	if vectorSearch == nil {
		log.Info().Msg("Initializing vector search service (cold start)")
		vectorSearch, err = services.NewVectorSearchService(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize vector search service")
			return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize search service: %v", err))
		}
	}

	// Load chunks if not initialized (happens on cold start)
	if !vectorSearch.IsInitialized() {
		loadStart := time.Now()
		log.Info().Msg("Loading knowledge base chunks (cold start)")
		if err := vectorSearch.LoadChunks(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to load chunks from DynamoDB")
			return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to load knowledge base: %v", err))
		}
		log.Info().Dur("load_duration_ms", time.Since(loadStart)).Msg("Knowledge base loaded successfully")
	}

	// Search for relevant chunks (top 5)
	searchStart := time.Now()
	log.Debug().Msg("Searching for relevant chunks")
	searchResults, err := vectorSearch.SearchByChannel(ctx, channelID, req.Question, 5)
	if err != nil {
		log.Error().Err(err).Msg("Vector search failed")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to search knowledge base: %v", err))
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
			channelID,
			req.Question,
			"I couldn't find any relevant information in this channel's videos to answer your question. The channel may not have any transcriptions uploaded yet.",
			[]models.MessageSource{},
		)

		if err := messageRepo.CreateMessage(ctx, message); err != nil {
			log.Error().Err(err).Msg("Failed to save empty message")
			return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to save message: %v", err))
		}

		log.Info().Str("message_id", message.GetID()).Msg("Returned empty answer")
		return response.JSON(http.StatusOK, SendMessageResponse{
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
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize repository: %v", err))
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
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize AI service: %v", err))
	}

	// Generate answer using Claude with RAG and selected prompt style
	generateStart := time.Now()
	log.Info().
		Int("context_chunk_count", len(contextChunks)).
		Str("prompt_style", string(promptStyle)).
		Any("chunks", contextChunks).
		Msg("Calling Bedrock to generate answer")
	answer, err := bedrockService.GenerateAnswerWithStyle(ctx, req.Question, contextChunks, promptStyle)
	if err != nil {
		log.Error().Err(err).Msg("Bedrock GenerateAnswer failed")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to generate answer: %v", err))
	}

	log.Info().
		Int("answer_length", len(answer)).
		Dur("generate_duration_ms", time.Since(generateStart)).
		Msg("Answer generated successfully")

	// Create and save message
	message := models.NewChatMessage(
		userID,
		channelID,
		req.Question,
		answer,
		sources,
	)

	log.Debug().Msg("Saving message to DynamoDB")
	if err := messageRepo.CreateMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Failed to save message to DynamoDB")
		return response.Error(http.StatusInternalServerError, fmt.Sprintf("Failed to save message: %v", err))
	}

	// Increment user's question count
	user.QuestionsUsedThisMonth++
	user.UpdatedAt = time.Now()
	if err := userRepo.SaveUser(ctx, &user); err != nil {
		log.Error().Err(err).Msg("Failed to update user question count")
		// Continue anyway, message was already saved
	} else {
		log.Info().
			Int("questions_used", user.QuestionsUsedThisMonth).
			Int("questions_limit", user.QuestionsLimit).
			Msg("Updated user question count")
	}

	duration := time.Since(start)
	log.Info().
		Str("message_id", message.GetID()).
		Int("source_count", len(sources)).
		Int("answer_length", len(answer)).
		Dur("total_duration_ms", duration).
		Msg("Send message request completed successfully")

	return response.JSON(http.StatusOK, SendMessageResponse{
		MessageID:   message.GetID(),
		Question:    message.Question,
		Answer:      message.Answer,
		Sources:     message.Sources,
		PromptStyle: string(promptStyle),
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func main() {
	lambda.Start(handler)
}
