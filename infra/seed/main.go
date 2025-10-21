package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
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

// Admin represents an admin user (DynamoDB record)
type Admin struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	Email     string `dynamodbav:"email"`
	CognitoID string `dynamodbav:"cognito_id"` // Store Cognito user ID instead of password
	CreatedAt string `dynamodbav:"created_at"`
}

func main() {
	// Load AWS config
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Create Cognito client
	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	// Get Cognito Admin Pool ID from environment
	adminPoolID := os.Getenv("COGNITO_ADMIN_POOL_ID")
	if adminPoolID == "" {
		log.Fatal("COGNITO_ADMIN_POOL_ID environment variable not set")
	}

	// Seed channels
	if err := seedChannels(ctx, dynamoClient); err != nil {
		log.Fatalf("failed to seed channels: %v", err)
	}

	// Seed admin user in Cognito and DynamoDB
	if err := seedAdmin(ctx, dynamoClient, cognitoClient, adminPoolID); err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}

	fmt.Println("\n✅ Seed data created successfully!")
	fmt.Println("📝 Admin credentials:")
	fmt.Println("   Email: admin@askly.com")
	fmt.Println("   Password: AdminPass123!")
	fmt.Println("   ⚠️  CHANGE THIS PASSWORD AFTER FIRST LOGIN!")
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

func seedAdmin(ctx context.Context, dynamoClient *dynamodb.Client, cognitoClient *cognitoidentityprovider.Client, adminPoolID string) error {
	email := "admin@askly.com"
	password := "AdminPass123!" // Meets Cognito admin requirements: min 12 chars, uppercase, lowercase, number, symbol

	fmt.Println("\n🔐 Creating admin user in Cognito...")

	// Create admin user in Cognito
	createUserInput := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(adminPoolID),
		Username:          aws.String(email),
		TemporaryPassword: aws.String(password),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
		MessageAction: types.MessageActionTypeSuppress, // Don't send welcome email
	}

	createUserOutput, err := cognitoClient.AdminCreateUser(ctx, createUserInput)
	if err != nil {
		return fmt.Errorf("failed to create admin in Cognito: %w", err)
	}

	// Extract Cognito user ID (sub)
	var cognitoUserID string
	for _, attr := range createUserOutput.User.Attributes {
		if aws.ToString(attr.Name) == "sub" {
			cognitoUserID = aws.ToString(attr.Value)
			break
		}
	}

	if cognitoUserID == "" {
		return fmt.Errorf("failed to get Cognito user ID")
	}

	fmt.Printf("✅ Created admin in Cognito (ID: %s)\n", cognitoUserID)

	// Set permanent password
	setPwdInput := &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(adminPoolID),
		Username:   aws.String(email),
		Password:   aws.String(password),
		Permanent:  true,
	}

	_, err = cognitoClient.AdminSetUserPassword(ctx, setPwdInput)
	if err != nil {
		return fmt.Errorf("failed to set permanent password: %w", err)
	}

	fmt.Println("✅ Set permanent password")

	// Create admin record in DynamoDB
	adminID := uuid.New().String()
	admin := Admin{
		PK:        fmt.Sprintf("ADMIN#%s", adminID),
		SK:        "PROFILE",
		Email:     email,
		CognitoID: cognitoUserID,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	item, err := attributevalue.MarshalMap(admin)
	if err != nil {
		return fmt.Errorf("failed to marshal admin: %w", err)
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("askly-admins"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put admin in DynamoDB: %w", err)
	}

	fmt.Printf("✅ Created admin record in DynamoDB (Admin ID: %s)\n", admin.PK)

	return nil
}
