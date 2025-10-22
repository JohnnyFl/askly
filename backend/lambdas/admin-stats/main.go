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

type StatsResponse struct {
	TotalChannels       int `json:"totalChannels"`
	TotalUsers          int `json:"totalUsers"`
	TotalQuestions      int `json:"totalQuestions"`
	TotalTranscriptions int `json:"totalTranscriptions"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize logger with request context
	ctx = logger.NewContextFromAPIGatewayRequest(ctx, request)
	log := logger.FromContext(ctx)

	start := time.Now()
	log.Info().Msg("Admin stats request received")

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

	log.Info().Str("email", email).Msg("Admin requesting stats")

	// Initialize repositories
	log.Debug().Msg("Initializing repositories")
	channelRepo, err := repository.NewChannelRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize channel repository")
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Count channels
	log.Debug().Msg("Counting channels")
	var channels []*models.Channel
	totalChannels := 0
	if err := channelRepo.ListChannels(ctx, &channels); err != nil {
		log.Error().Err(err).Msg("Failed to count channels")
	} else {
		totalChannels = len(channels)
		log.Debug().Int("total_channels", totalChannels).Msg("Channels counted")
	}

	// Count users by scanning the users table
	log.Debug().Msg("Counting users")
	totalUsers := 0
	var users []map[string]interface{}
	db, err := repository.NewDynamoDBRepository(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize DB repository")
	} else {
		if err := db.Scan(ctx, "askly-users", &users); err != nil {
			log.Error().Err(err).Str("table", "askly-users").Msg("Failed to scan users table")
		} else {
			totalUsers = len(users)
			log.Debug().Int("total_users", totalUsers).Msg("Users counted")
		}
	}

	// Count transcriptions by scanning transcriptions table
	log.Debug().Msg("Counting transcriptions")
	totalTranscriptions := 0
	var transcriptions []map[string]interface{}
	if err := db.Scan(ctx, "askly-transcriptions", &transcriptions); err != nil {
		log.Error().Err(err).Str("table", "askly-transcriptions").Msg("Failed to scan transcriptions table")
	} else {
		totalTranscriptions = len(transcriptions)
		log.Debug().Int("total_transcriptions", totalTranscriptions).Msg("Transcriptions counted")
	}

	// Count messages (questions) by scanning messages table
	log.Debug().Msg("Counting messages")
	totalQuestions := 0
	var messages []map[string]interface{}
	if err := db.Scan(ctx, "askly-messages", &messages); err != nil {
		log.Error().Err(err).Str("table", "askly-messages").Msg("Failed to scan messages table")
	} else {
		totalQuestions = len(messages)
		log.Debug().Int("total_questions", totalQuestions).Msg("Messages counted")
	}

	// Log final stats summary
	duration := time.Since(start)
	log.Info().
		Int("total_channels", totalChannels).
		Int("total_users", totalUsers).
		Int("total_questions", totalQuestions).
		Int("total_transcriptions", totalTranscriptions).
		Dur("duration_ms", duration).
		Msg("Stats retrieved successfully")

	return response(http.StatusOK, StatsResponse{
		TotalChannels:       totalChannels,
		TotalUsers:          totalUsers,
		TotalQuestions:      totalQuestions,
		TotalTranscriptions: totalTranscriptions,
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
