package repository

import (
	"context"
	"os"
)

const messagesTableEnv = "DYNAMODB_MESSAGES_TABLE"

// MessageRepository provides message-specific operations
type MessageRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(ctx context.Context) (*MessageRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, err
	}

	tableName := os.Getenv(messagesTableEnv)
	if tableName == "" {
		tableName = "askly-messages" // Default fallback
	}

	return &MessageRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateMessage creates a new chat message
func (r *MessageRepository) CreateMessage(ctx context.Context, message interface{}) error {
	return r.db.PutItem(ctx, r.tableName, message)
}

// GetMessageByID gets a message by ID
func (r *MessageRepository) GetMessageByID(ctx context.Context, messageID string, result interface{}) error {
	return r.db.GetItem(ctx, r.tableName, messageID, "METADATA", result)
}

// ListMessagesByUser lists messages for a user
func (r *MessageRepository) ListMessagesByUser(ctx context.Context, userID string, results interface{}) error {
	return r.db.QueryByGSI(ctx, r.tableName, "user_id-created_at-index", "user_id", userID, results)
}

// ListMessagesByChannel lists messages for a channel
func (r *MessageRepository) ListMessagesByChannel(ctx context.Context, channelID string, results interface{}) error {
	return r.db.QueryByGSI(ctx, r.tableName, "channel_id-created_at-index", "channel_id", channelID, results)
}

// DeleteMessage deletes a message
func (r *MessageRepository) DeleteMessage(ctx context.Context, messageID string) error {
	return r.db.DeleteItem(ctx, r.tableName, messageID, "METADATA")
}
