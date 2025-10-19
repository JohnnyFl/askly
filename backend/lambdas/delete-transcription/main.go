package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"askly/backend/shared/auth"
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
	// Verify JWT token - only admins can delete transcriptions
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

	// Only admins can delete transcriptions
	if !claims.IsAdmin() {
		return response(http.StatusForbidden, ErrorResponse{Error: "Admin access required"})
	}

	// Get channel ID and transcription ID from path parameters
	channelID := request.PathParameters["id"]
	transcriptionID := request.PathParameters["transcriptionId"]

	if channelID == "" || transcriptionID == "" {
		return response(http.StatusBadRequest, ErrorResponse{Error: "Channel ID and Transcription ID are required"})
	}

	transcriptionPK := "TRANSCRIPTION#" + transcriptionID
	log.Printf("INFO: Admin %s deleting transcription %s from channel %s", claims.UserID, transcriptionPK, channelID)

	// Initialize repositories
	transcriptionRepo, err := repository.NewTranscriptionRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize transcription repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	chunkRepo, err := repository.NewChunkRepository(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to initialize chunk repository: %v", err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error"})
	}

	// Delete all chunks associated with this transcription
	var chunks []*models.TranscriptionChunk
	if err := chunkRepo.ListChunksByTranscription(ctx, transcriptionPK, &chunks); err != nil {
		log.Printf("WARN: Failed to list chunks for transcription %s: %v", transcriptionPK, err)
	} else {
		log.Printf("INFO: Deleting %d chunks for transcription %s", len(chunks), transcriptionPK)
		for _, chunk := range chunks {
			if err := chunkRepo.DeleteChunk(ctx, chunk.PK); err != nil {
				log.Printf("ERROR: Failed to delete chunk %s: %v", chunk.PK, err)
			}
		}
	}

	// Delete the transcription itself
	if err := transcriptionRepo.DeleteTranscription(ctx, transcriptionPK); err != nil {
		log.Printf("ERROR: Failed to delete transcription %s: %v", transcriptionPK, err)
		return response(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete transcription"})
	}

	log.Printf("INFO: Successfully deleted transcription %s", transcriptionPK)

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
