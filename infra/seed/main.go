package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Channel represents a YouTube channel
type Channel struct {
	PK               string `dynamodbav:"PK"`
	SK               string `dynamodbav:"SK"`
	Name             string `dynamodbav:"name"`
	Description      string `dynamodbav:"description"`
	ThumbnailURL     string `dynamodbav:"thumbnail_url"`
	YoutubeURL       string `dynamodbav:"youtube_url"`
	YoutubeChannelID string `dynamodbav:"youtube_channel_id"`
	CreatedAt        string `dynamodbav:"created_at"`
}

// Admin represents an admin user
type Admin struct {
	PK           string `dynamodbav:"PK"`
	SK           string `dynamodbav:"SK"`
	Email        string `dynamodbav:"email"`
	PasswordHash string `dynamodbav:"password_hash"`
	CreatedAt    string `dynamodbav:"created_at"`
}

func main() {
	// Load AWS config
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DynamoDB client
	client := dynamodb.NewFromConfig(cfg)

	// Seed channels
	if err := seedChannels(ctx, client); err != nil {
		log.Fatalf("failed to seed channels: %v", err)
	}

	// Seed admin user
	if err := seedAdmin(ctx, client); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	fmt.Println("✅ Seed data created successfully!")
}

func seedChannels(ctx context.Context, client *dynamodb.Client) error {
	channels := []Channel{
		{
			PK:               fmt.Sprintf("CHANNEL#%s", uuid.New().String()),
			SK:               "METADATA",
			Name:             "Y Combinator",
			Description:      "Y Combinator created a new model for funding early stage startups. Twice a year we invest in a large number of startups. Learn about startups, entrepreneurship, and more.",
			ThumbnailURL:     "https://yt3.googleusercontent.com/ytc/AIdro_m2R0FWTlPPgjLJWnKEIFVmqZEQpUW6xXhE1VjxRQ=s176-c-k-c0x00ffffff-no-rj",
			YoutubeURL:       "https://www.youtube.com/@ycombinator",
			YoutubeChannelID: "UCcefcZRL2oaA_uBNeo5UOWg",
			CreatedAt:        time.Now().Format(time.RFC3339),
		},
		{
			PK:               fmt.Sprintf("CHANNEL#%s", uuid.New().String()),
			SK:               "METADATA",
			Name:             "Startup Grind",
			Description:      "Startup Grind is the world's largest startup community. We educate, inspire, and connect entrepreneurs through local and global events, online platforms, and our media company.",
			ThumbnailURL:     "https://yt3.googleusercontent.com/ytc/AIdro_kJvPZIyqLKAXN5aRGxGJ7_Oj9mYDHG_QQ9hvLv=s176-c-k-c0x00ffffff-no-rj",
			YoutubeURL:       "https://www.youtube.com/@StartupGrind",
			YoutubeChannelID: "UCBkfV8-N4nw8c4WV3rKxTwg",
			CreatedAt:        time.Now().Format(time.RFC3339),
		},
	}

	for _, channel := range channels {
		item, err := attributevalue.MarshalMap(channel)
		if err != nil {
			return fmt.Errorf("failed to marshal channel: %w", err)
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("askly-channels"),
			Item:      item,
		})
		if err != nil {
			return fmt.Errorf("failed to put channel: %w", err)
		}

		fmt.Printf("✅ Created channel: %s (ID: %s)\n", channel.Name, channel.PK)
	}

	return nil
}

func seedAdmin(ctx context.Context, client *dynamodb.Client) error {
	// Hash password
	password := "admin123" // Default password - should be changed after first login
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminID := uuid.New().String()
	admin := Admin{
		PK:           fmt.Sprintf("ADMIN#%s", adminID),
		SK:           "PROFILE",
		Email:        "admin@askly.com",
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	item, err := attributevalue.MarshalMap(admin)
	if err != nil {
		return fmt.Errorf("failed to marshal admin: %w", err)
	}

	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("askly-admins"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put admin: %w", err)
	}

	fmt.Printf("✅ Created admin user: %s\n", admin.Email)
	fmt.Printf("   Password: %s (change this after first login!)\n", password)
	fmt.Printf("   Admin ID: %s\n", admin.PK)

	return nil
}
