package repository

import (
	"context"
	"os"
)

const (
	transcriptionsTableEnv = "DYNAMODB_TRANSCRIPTIONS_TABLE"
	chunksTableEnv         = "DYNAMODB_CHUNKS_TABLE"
)

// TranscriptionRepository provides transcription-specific operations
type TranscriptionRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewTranscriptionRepository creates a new transcription repository
func NewTranscriptionRepository(ctx context.Context) (*TranscriptionRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, err
	}

	tableName := os.Getenv(transcriptionsTableEnv)
	if tableName == "" {
		tableName = "askly-transcriptions" // Default fallback
	}

	return &TranscriptionRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateTranscription creates a new transcription
func (r *TranscriptionRepository) CreateTranscription(ctx context.Context, transcription interface{}) error {
	return r.db.PutItem(ctx, r.tableName, transcription)
}

// GetTranscriptionByID gets a transcription by ID
func (r *TranscriptionRepository) GetTranscriptionByID(ctx context.Context, transcriptionID string, result interface{}) error {
	return r.db.GetItem(ctx, r.tableName, transcriptionID, "METADATA", result)
}

// ListTranscriptionsByChannel lists transcriptions for a channel
func (r *TranscriptionRepository) ListTranscriptionsByChannel(ctx context.Context, channelID string, results interface{}) error {
	return r.db.QueryByGSI(ctx, r.tableName, "channel_id-uploaded_at-index", "channel_id", channelID, results)
}

// DeleteTranscription deletes a transcription
func (r *TranscriptionRepository) DeleteTranscription(ctx context.Context, transcriptionID string) error {
	return r.db.DeleteItem(ctx, r.tableName, transcriptionID, "METADATA")
}

// ChunkRepository provides chunk-specific operations
type ChunkRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewChunkRepository creates a new chunk repository
func NewChunkRepository(ctx context.Context) (*ChunkRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, err
	}

	tableName := os.Getenv(chunksTableEnv)
	if tableName == "" {
		tableName = "askly-chunks" // Default fallback
	}

	return &ChunkRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateChunk creates a new chunk
func (r *ChunkRepository) CreateChunk(ctx context.Context, chunk interface{}) error {
	return r.db.PutItem(ctx, r.tableName, chunk)
}

// GetChunkByID gets a chunk by ID
func (r *ChunkRepository) GetChunkByID(ctx context.Context, chunkID string, result interface{}) error {
	return r.db.GetItem(ctx, r.tableName, chunkID, "METADATA", result)
}

// ListChunksByChannel lists all chunks for a channel
func (r *ChunkRepository) ListChunksByChannel(ctx context.Context, channelID string, results interface{}) error {
	return r.db.QueryByGSI(ctx, r.tableName, "channel_id-index", "channel_id", channelID, results)
}

// ListChunksByTranscription lists chunks for a transcription
func (r *ChunkRepository) ListChunksByTranscription(ctx context.Context, transcriptionID string, results interface{}) error {
	return r.db.QueryByGSI(ctx, r.tableName, "transcription_id-chunk_index-index", "transcription_id", transcriptionID, results)
}

// ScanAllChunks scans all chunks (for loading into Lambda memory for vector search)
func (r *ChunkRepository) ScanAllChunks(ctx context.Context, results interface{}) error {
	return r.db.Scan(ctx, r.tableName, results)
}

// DeleteChunk deletes a chunk
func (r *ChunkRepository) DeleteChunk(ctx context.Context, chunkID string) error {
	return r.db.DeleteItem(ctx, r.tableName, chunkID, "METADATA")
}
