package models

import "time"

// Transcription represents a video transcription
type Transcription struct {
	PK         string    `dynamodbav:"PK" json:"id"`
	SK         string    `dynamodbav:"SK" json:"-"`
	ChannelID  string    `dynamodbav:"channel_id" json:"channel_id"`
	VideoTitle string    `dynamodbav:"video_title" json:"video_title"`
	VideoURL   string    `dynamodbav:"video_url" json:"video_url"`
	VideoID    string    `dynamodbav:"video_id" json:"video_id"`
	FullText   string    `dynamodbav:"full_text" json:"full_text"`
	UploadedBy string    `dynamodbav:"uploaded_by" json:"uploaded_by"`
	UploadedAt time.Time `dynamodbav:"uploaded_at" json:"uploaded_at"`
	Status     string    `dynamodbav:"status" json:"status"` // "completed", "processing", "failed"
}

// TranscriptionChunk represents a chunk of a transcription with embedding
type TranscriptionChunk struct {
	PK              string    `dynamodbav:"PK" json:"id"`
	SK              string    `dynamodbav:"SK" json:"-"`
	TranscriptionID string    `dynamodbav:"transcription_id" json:"transcription_id"`
	ChannelID       string    `dynamodbav:"channel_id" json:"channel_id"`
	ChunkText       string    `dynamodbav:"chunk_text" json:"chunk_text"`
	ChunkIndex      int       `dynamodbav:"chunk_index" json:"chunk_index"`
	Embedding       []byte    `dynamodbav:"embedding" json:"-"` // Binary representation of []float32
	CreatedAt       time.Time `dynamodbav:"created_at" json:"created_at"`
}

// NewTranscription creates a new Transcription with generated ID
func NewTranscription(channelID, videoTitle, videoURL, videoID, fullText, uploadedBy string) *Transcription {
	return &Transcription{
		PK:         "TRANSCRIPTION#" + GenerateID(),
		SK:         "METADATA",
		ChannelID:  channelID,
		VideoTitle: videoTitle,
		VideoURL:   videoURL,
		VideoID:    videoID,
		FullText:   fullText,
		UploadedBy: uploadedBy,
		UploadedAt: time.Now(),
		Status:     "completed", // Default to completed
	}
}

// NewTranscriptionChunk creates a new TranscriptionChunk with generated ID
func NewTranscriptionChunk(transcriptionID, channelID, chunkText string, chunkIndex int, embedding []byte) *TranscriptionChunk {
	return &TranscriptionChunk{
		PK:              "CHUNK#" + GenerateID(),
		SK:              "METADATA",
		TranscriptionID: transcriptionID,
		ChannelID:       channelID,
		ChunkText:       chunkText,
		ChunkIndex:      chunkIndex,
		Embedding:       embedding,
		CreatedAt:       time.Now(),
	}
}

// GetID extracts the ID from PK
func (t *Transcription) GetID() string {
	if len(t.PK) > 14 {
		return t.PK[14:] // Remove "TRANSCRIPTION#" prefix
	}
	return t.PK
}

// GetID extracts the ID from PK
func (tc *TranscriptionChunk) GetID() string {
	if len(tc.PK) > 6 {
		return tc.PK[6:] // Remove "CHUNK#" prefix
	}
	return tc.PK
}
