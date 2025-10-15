package models

import "time"

// ChatMessage represents a user question and AI answer
type ChatMessage struct {
	PK        string          `dynamodbav:"PK" json:"id"`
	SK        string          `dynamodbav:"SK" json:"-"`
	UserID    string          `dynamodbav:"user_id" json:"user_id"`
	ChannelID string          `dynamodbav:"channel_id" json:"channel_id"`
	Question  string          `dynamodbav:"question" json:"question"`
	Answer    string          `dynamodbav:"answer" json:"answer"`
	Sources   []MessageSource `dynamodbav:"sources" json:"sources"`
	CreatedAt time.Time       `dynamodbav:"created_at" json:"created_at"`
}

// MessageSource represents the source of information for an answer
type MessageSource struct {
	VideoTitle     string  `dynamodbav:"video_title" json:"video_title"`
	VideoURL       string  `dynamodbav:"video_url" json:"video_url"`
	ChunkIndex     int     `dynamodbav:"chunk_index" json:"chunk_index"`
	RelevanceScore float64 `dynamodbav:"relevance_score" json:"relevance_score"`
}

// NewChatMessage creates a new ChatMessage with generated ID
func NewChatMessage(userID, channelID, question, answer string, sources []MessageSource) *ChatMessage {
	return &ChatMessage{
		PK:        "MESSAGE#" + GenerateID(),
		SK:        "METADATA",
		UserID:    userID,
		ChannelID: channelID,
		Question:  question,
		Answer:    answer,
		Sources:   sources,
		CreatedAt: time.Now(),
	}
}

// GetID extracts the ID from PK
func (m *ChatMessage) GetID() string {
	if len(m.PK) > 8 {
		return m.PK[8:] // Remove "MESSAGE#" prefix
	}
	return m.PK
}
