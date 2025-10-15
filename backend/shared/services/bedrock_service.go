package services

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockService provides AI capabilities using AWS Bedrock
type BedrockService struct {
	client *bedrockruntime.Client
}

// NewBedrockService creates a new Bedrock service
func NewBedrockService(ctx context.Context) (*BedrockService, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockService{
		client: client,
	}, nil
}

// TitanEmbedRequest represents the request body for Titan Embed Text v2
type TitanEmbedRequest struct {
	InputText string `json:"inputText"`
}

// TitanEmbedResponse represents the response from Titan Embed Text v2
type TitanEmbedResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

// GenerateEmbedding generates embeddings using Amazon Titan Embed Text v2
func (s *BedrockService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Prepare request body
	requestBody := TitanEmbedRequest{
		InputText: text,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call Bedrock API
	response, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("amazon.titan-embed-text-v2:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestJSON,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to invoke Titan Embed model: %w", err)
	}

	// Parse response
	var embedResponse TitanEmbedResponse
	if err := json.Unmarshal(response.Body, &embedResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding response: %w", err)
	}

	return embedResponse.Embedding, nil
}

// ClaudeMessage represents a message in the Claude API
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeRequest represents the request body for Claude 3 Haiku
type ClaudeRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxTokens        int             `json:"max_tokens"`
	Messages         []ClaudeMessage `json:"messages"`
	Temperature      float64         `json:"temperature,omitempty"`
}

// ClaudeResponse represents the response from Claude 3 Haiku
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// GenerateAnswer generates an answer using Claude 3 Haiku with RAG context
func (s *BedrockService) GenerateAnswer(ctx context.Context, question string, contextChunks []string) (string, error) {
	// Construct the prompt with context
	contextText := strings.Join(contextChunks, "\n\n---\n\n")

	prompt := fmt.Sprintf(`You are a helpful AI assistant that answers questions based on the provided context from YouTube video transcriptions.

Context from video transcriptions:
%s

Question: %s

Please provide a clear and concise answer based on the context above. If the context doesn't contain enough information to answer the question, say so.`, contextText, question)

	// Prepare request body
	requestBody := ClaudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call Bedrock API for Claude 3 Haiku
	response, err := s.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        requestJSON,
	})

	if err != nil {
		return "", fmt.Errorf("failed to invoke Claude model: %w", err)
	}

	// Parse response
	var claudeResponse ClaudeResponse
	if err := json.Unmarshal(response.Body, &claudeResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal Claude response: %w", err)
	}

	// Extract text from response
	if len(claudeResponse.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	return claudeResponse.Content[0].Text, nil
}

// EmbeddingToBytes converts a float32 embedding to bytes for DynamoDB storage
func EmbeddingToBytes(embedding []float32) []byte {
	bytes := make([]byte, len(embedding)*4) // 4 bytes per float32
	for i, val := range embedding {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(val))
	}
	return bytes
}

// BytesToEmbedding converts bytes back to float32 embedding from DynamoDB
func BytesToEmbedding(bytes []byte) ([]float32, error) {
	if len(bytes)%4 != 0 {
		return nil, fmt.Errorf("invalid embedding bytes length: %d", len(bytes))
	}

	embedding := make([]float32, len(bytes)/4)
	for i := range embedding {
		bits := binary.LittleEndian.Uint32(bytes[i*4:])
		embedding[i] = math.Float32frombits(bits)
	}
	return embedding, nil
}

// CosineSimilarity calculates cosine similarity between two embeddings
func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("embedding dimensions don't match: %d vs %d", len(a), len(b))
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("zero vector detected")
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB)))), nil
}
