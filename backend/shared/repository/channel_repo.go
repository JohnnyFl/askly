package repository

import (
	"context"
	"os"
)

const channelsTableEnv = "DYNAMODB_CHANNELS_TABLE"

// ChannelRepository provides channel-specific operations
type ChannelRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewChannelRepository creates a new channel repository
func NewChannelRepository(ctx context.Context) (*ChannelRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, err
	}

	tableName := os.Getenv(channelsTableEnv)
	if tableName == "" {
		tableName = "askly-channels" // Default fallback
	}

	return &ChannelRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateChannel creates a new channel
func (r *ChannelRepository) CreateChannel(ctx context.Context, channel interface{}) error {
	return r.db.PutItem(ctx, r.tableName, channel)
}

// GetChannelByID gets a channel by ID
func (r *ChannelRepository) GetChannelByID(ctx context.Context, channelID string, result interface{}) error {
	return r.db.GetItem(ctx, r.tableName, channelID, "METADATA", result)
}

// ListChannels lists all channels
func (r *ChannelRepository) ListChannels(ctx context.Context, results interface{}) error {
	return r.db.Scan(ctx, r.tableName, results)
}

// UpdateChannel updates channel attributes
func (r *ChannelRepository) UpdateChannel(ctx context.Context, channelID string, updates map[string]interface{}) error {
	return r.db.UpdateItem(ctx, r.tableName, channelID, "METADATA", updates)
}

// DeleteChannel deletes a channel
func (r *ChannelRepository) DeleteChannel(ctx context.Context, channelID string) error {
	return r.db.DeleteItem(ctx, r.tableName, channelID, "METADATA")
}
