package services

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"askly/backend/shared/models"
	"askly/backend/shared/repository"
)

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Chunk      *models.TranscriptionChunk
	Similarity float32
}

// VectorSearchService provides in-memory vector search functionality
type VectorSearchService struct {
	chunkRepo      *repository.ChunkRepository
	bedrockService *BedrockService
	chunks         []*models.TranscriptionChunk
	embeddings     [][]float32
	mu             sync.RWMutex
	initialized    bool
}

// NewVectorSearchService creates a new vector search service
func NewVectorSearchService(ctx context.Context) (*VectorSearchService, error) {
	chunkRepo, err := repository.NewChunkRepository(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk repository: %w", err)
	}

	bedrockService, err := NewBedrockService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bedrock service: %w", err)
	}

	return &VectorSearchService{
		chunkRepo:      chunkRepo,
		bedrockService: bedrockService,
		chunks:         make([]*models.TranscriptionChunk, 0),
		embeddings:     make([][]float32, 0),
		initialized:    false,
	}, nil
}

// LoadChunks loads all chunks from DynamoDB into memory
func (s *VectorSearchService) LoadChunks(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var chunks []*models.TranscriptionChunk
	if err := s.chunkRepo.ScanAllChunks(ctx, &chunks); err != nil {
		return fmt.Errorf("failed to scan chunks: %w", err)
	}

	// Convert embeddings from []byte to []float32
	embeddings := make([][]float32, len(chunks))
	for i, chunk := range chunks {
		if len(chunk.Embedding) == 0 {
			return fmt.Errorf("chunk %s has no embedding", chunk.PK)
		}

		embedding, err := BytesToEmbedding(chunk.Embedding)
		if err != nil {
			return fmt.Errorf("failed to convert embedding for chunk %s: %w", chunk.PK, err)
		}
		embeddings[i] = embedding
	}

	s.chunks = chunks
	s.embeddings = embeddings
	s.initialized = true

	return nil
}

// Search performs vector similarity search and returns top K results
func (s *VectorSearchService) Search(ctx context.Context, query string, topK int) ([]*SearchResult, error) {
	// Check if service is initialized
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		// Initialize if not already done
		if err := s.LoadChunks(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize vector search: %w", err)
		}
		s.mu.RLock()
	}
	defer s.mu.RUnlock()

	if len(s.chunks) == 0 {
		return []*SearchResult{}, nil
	}

	// Generate embedding for query
	queryEmbedding, err := s.bedrockService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Calculate similarities for all chunks
	results := make([]*SearchResult, len(s.chunks))
	for i := range s.chunks {
		similarity, err := CosineSimilarity(queryEmbedding, s.embeddings[i])
		if err != nil {
			return nil, fmt.Errorf("failed to calculate similarity for chunk %d: %w", i, err)
		}

		results[i] = &SearchResult{
			Chunk:      s.chunks[i],
			Similarity: similarity,
		}
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Return top K results
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// SearchByChannel performs vector search filtered by channel ID
func (s *VectorSearchService) SearchByChannel(ctx context.Context, channelID string, query string, topK int) ([]*SearchResult, error) {
	// Check if service is initialized
	s.mu.RLock()
	if !s.initialized {
		s.mu.RUnlock()
		// Initialize if not already done
		if err := s.LoadChunks(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize vector search: %w", err)
		}
		s.mu.RLock()
	}
	defer s.mu.RUnlock()

	if len(s.chunks) == 0 {
		return []*SearchResult{}, nil
	}

	// Generate embedding for query
	queryEmbedding, err := s.bedrockService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Calculate similarities only for chunks from the specified channel
	var results []*SearchResult
	for i := range s.chunks {
		if s.chunks[i].ChannelID != channelID {
			continue
		}

		similarity, err := CosineSimilarity(queryEmbedding, s.embeddings[i])
		if err != nil {
			return nil, fmt.Errorf("failed to calculate similarity for chunk %d: %w", i, err)
		}

		results = append(results, &SearchResult{
			Chunk:      s.chunks[i],
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Return top K results
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// RefreshChunks reloads chunks from DynamoDB (call after new transcriptions are added)
func (s *VectorSearchService) RefreshChunks(ctx context.Context) error {
	return s.LoadChunks(ctx)
}

// GetStats returns statistics about loaded chunks
func (s *VectorSearchService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channelCounts := make(map[string]int)
	for _, chunk := range s.chunks {
		channelCounts[chunk.ChannelID]++
	}

	return map[string]interface{}{
		"initialized":          s.initialized,
		"total_chunks":         len(s.chunks),
		"chunks_per_channel":   channelCounts,
		"embedding_dimensions": len(s.embeddings[0]),
	}
}

// IsInitialized returns whether the service has been initialized
func (s *VectorSearchService) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}
