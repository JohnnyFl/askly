package models

import "time"

// QuestionUsage tracks individual question usage for analytics
type QuestionUsage struct {
	PK           string    `dynamodbav:"PK" json:"id"` // usage_id
	SK           string    `dynamodbav:"SK" json:"-"`  // "USAGE"
	UserID       string    `dynamodbav:"user_id" json:"user_id"`
	ChannelID    string    `dynamodbav:"channel_id" json:"channel_id"`
	QuestionText string    `dynamodbav:"question_text,omitempty" json:"question_text,omitempty"`
	MonthYear    string    `dynamodbav:"month_year" json:"month_year"` // "2025-01" format
	TTL          int64     `dynamodbav:"ttl" json:"-"`                 // Auto-delete after 90 days
	CreatedAt    time.Time `dynamodbav:"created_at" json:"created_at"`
}

// NewQuestionUsage creates a new QuestionUsage record with automatic TTL
func NewQuestionUsage(userID, channelID, questionText string) *QuestionUsage {
	now := time.Now()

	// Set TTL to 90 days from now (for automatic cleanup)
	ttl := now.AddDate(0, 0, 90).Unix()

	return &QuestionUsage{
		PK:           GenerateID(),
		SK:           "USAGE",
		UserID:       userID,
		ChannelID:    channelID,
		QuestionText: questionText,
		MonthYear:    now.Format("2006-01"), // YYYY-MM format
		TTL:          ttl,
		CreatedAt:    now,
	}
}
