package models

import "time"

// Channel represents a YouTube channel
type Channel struct {
	PK               string    `dynamodbav:"PK" json:"id"`
	SK               string    `dynamodbav:"SK" json:"-"`
	Name             string    `dynamodbav:"name" json:"name"`
	Description      string    `dynamodbav:"description" json:"description"`
	ThumbnailURL     string    `dynamodbav:"thumbnail_url" json:"thumbnail_url"`
	YoutubeURL       string    `dynamodbav:"youtube_url" json:"youtube_url"`
	YoutubeChannelID string    `dynamodbav:"youtube_channel_id" json:"youtube_channel_id"`
	CreatedAt        time.Time `dynamodbav:"created_at" json:"created_at"`
}

// NewChannel creates a new Channel with generated ID
func NewChannel(name, description, thumbnailURL, youtubeURL, youtubeChannelID string) *Channel {
	return &Channel{
		PK:               GenerateID(),
		SK:               "METADATA",
		Name:             name,
		Description:      description,
		ThumbnailURL:     thumbnailURL,
		YoutubeURL:       youtubeURL,
		YoutubeChannelID: youtubeChannelID,
		CreatedAt:        time.Now(),
	}
}

// GetID returns the ID (PK is now the ID directly)
func (c *Channel) GetID() string {
	return c.PK
}
