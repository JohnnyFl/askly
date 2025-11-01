package main

import (
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// ===== Configuration =====
		// Paddle configuration - reads from Pulumi config OR environment variables
		cfg := config.New(ctx, "")

		// Helper function to get config with env var fallback
		getConfig := func(key, envVar, defaultValue string) string {
			// Try Pulumi config first
			if val := cfg.Get(key); val != "" {
				return val
			}
			// Fall back to environment variable
			if val := os.Getenv(envVar); val != "" {
				return val
			}
			// Use default
			return defaultValue
		}

		// Paddle API settings
		paddleAPIKey := getConfig("paddle:apiKey", "PADDLE_API_KEY", "paddle_api_key_placeholder")
		paddleWebhookSecret := getConfig("paddle:webhookSecret", "PADDLE_WEBHOOK_SECRET", "paddle_webhook_secret_placeholder")
		paddleEnvironment := getConfig("paddle:environment", "PADDLE_ENVIRONMENT", "sandbox")

		// Paddle Price IDs
		paddleProMonthlyPriceID := getConfig("paddle:proMonthlyPriceId", "PADDLE_PRO_MONTHLY_PRICE_ID", "pri_placeholder_pro_monthly")
		paddleProAnnualPriceID := getConfig("paddle:proAnnualPriceId", "PADDLE_PRO_ANNUAL_PRICE_ID", "pri_placeholder_pro_annual")
		paddlePremiumMonthlyPriceID := getConfig("paddle:premiumMonthlyPriceId", "PADDLE_PREMIUM_MONTHLY_PRICE_ID", "pri_placeholder_premium_monthly")
		paddlePremiumAnnualPriceID := getConfig("paddle:premiumAnnualPriceId", "PADDLE_PREMIUM_ANNUAL_PRICE_ID", "pri_placeholder_premium_annual")

		// ===== DynamoDB Tables =====
		// Create Users table
		usersTable, err := dynamodb.NewTable(ctx, "askly-users", &dynamodb.TableArgs{
			Name:        pulumi.String("askly-users"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("PK"),
			RangeKey:    pulumi.String("SK"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("SK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("email"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("paddle_customer_id"),
					Type: pulumi.String("S"),
				},
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("email-index"),
					HashKey:        pulumi.String("email"),
					ProjectionType: pulumi.String("ALL"),
				},
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("paddle-customer-index"),
					HashKey:        pulumi.String("paddle_customer_id"),
					ProjectionType: pulumi.String("ALL"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-users"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Admins table
		adminsTable, err := dynamodb.NewTable(ctx, "askly-admins", &dynamodb.TableArgs{
			Name:        pulumi.String("askly-admins"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("PK"),
			RangeKey:    pulumi.String("SK"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("SK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("email"),
					Type: pulumi.String("S"),
				},
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("email-index"),
					HashKey:        pulumi.String("email"),
					ProjectionType: pulumi.String("ALL"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-admins"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Channels table
		channelsTable, err := dynamodb.NewTable(ctx, "askly-channels", &dynamodb.TableArgs{
			Name:        pulumi.String("askly-channels"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("PK"),
			RangeKey:    pulumi.String("SK"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("SK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("youtube_channel_id"),
					Type: pulumi.String("S"),
				},
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("youtube_channel_id-index"),
					HashKey:        pulumi.String("youtube_channel_id"),
					ProjectionType: pulumi.String("ALL"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-channels"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Transcriptions table
		transcriptionsTable, err := dynamodb.NewTable(ctx, "askly-transcriptions", &dynamodb.TableArgs{
			Name:        pulumi.String("askly-transcriptions"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("PK"),
			RangeKey:    pulumi.String("SK"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("SK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("channel_id"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("uploaded_at"),
					Type: pulumi.String("S"),
				},
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("channel_id-uploaded_at-index"),
					HashKey:        pulumi.String("channel_id"),
					RangeKey:       pulumi.String("uploaded_at"),
					ProjectionType: pulumi.String("ALL"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-transcriptions"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Chunks table
		chunksTable, err := dynamodb.NewTable(ctx, "askly-chunks", &dynamodb.TableArgs{
			Name:        pulumi.String("askly-chunks"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("PK"),
			RangeKey:    pulumi.String("SK"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("SK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("channel_id"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("transcription_id"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("chunk_index"),
					Type: pulumi.String("N"),
				},
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("channel_id-index"),
					HashKey:        pulumi.String("channel_id"),
					ProjectionType: pulumi.String("ALL"),
				},
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("transcription_id-chunk_index-index"),
					HashKey:        pulumi.String("transcription_id"),
					RangeKey:       pulumi.String("chunk_index"),
					ProjectionType: pulumi.String("ALL"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-chunks"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Messages table
		messagesTable, err := dynamodb.NewTable(ctx, "askly-messages", &dynamodb.TableArgs{
			Name:        pulumi.String("askly-messages"),
			BillingMode: pulumi.String("PAY_PER_REQUEST"),
			HashKey:     pulumi.String("PK"),
			RangeKey:    pulumi.String("SK"),
			Attributes: dynamodb.TableAttributeArray{
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("PK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("SK"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("user_id"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("channel_id"),
					Type: pulumi.String("S"),
				},
				&dynamodb.TableAttributeArgs{
					Name: pulumi.String("created_at"),
					Type: pulumi.String("S"),
				},
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("user_id-created_at-index"),
					HashKey:        pulumi.String("user_id"),
					RangeKey:       pulumi.String("created_at"),
					ProjectionType: pulumi.String("ALL"),
				},
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("channel_id-created_at-index"),
					HashKey:        pulumi.String("channel_id"),
					RangeKey:       pulumi.String("created_at"),
					ProjectionType: pulumi.String("ALL"),
				},
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-messages"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// ===== Cognito User Pool =====
		// Create User Pool for regular users
		userPool, err := cognito.NewUserPool(ctx, "askly-user-pool", &cognito.UserPoolArgs{
			Name: pulumi.String("askly-users"),
			PasswordPolicy: &cognito.UserPoolPasswordPolicyArgs{
				MinimumLength:    pulumi.Int(8),
				RequireLowercase: pulumi.Bool(true),
				RequireNumbers:   pulumi.Bool(true),
				RequireSymbols:   pulumi.Bool(false),
				RequireUppercase: pulumi.Bool(true),
			},
			AutoVerifiedAttributes: pulumi.StringArray{
				pulumi.String("email"),
			},
			UsernameAttributes: pulumi.StringArray{
				pulumi.String("email"),
			},
			Schemas: cognito.UserPoolSchemaArray{
				&cognito.UserPoolSchemaArgs{
					Name:                   pulumi.String("email"),
					AttributeDataType:      pulumi.String("String"),
					Required:               pulumi.Bool(true),
					Mutable:                pulumi.Bool(true),
					DeveloperOnlyAttribute: pulumi.Bool(false),
				},
			},
			EmailConfiguration: &cognito.UserPoolEmailConfigurationArgs{
				EmailSendingAccount: pulumi.String("COGNITO_DEFAULT"),
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-user-pool"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create User Pool Client for frontend
		userPoolClient, err := cognito.NewUserPoolClient(ctx, "askly-user-pool-client", &cognito.UserPoolClientArgs{
			Name:       pulumi.String("askly-web-client"),
			UserPoolId: userPool.ID(),
			ExplicitAuthFlows: pulumi.StringArray{
				pulumi.String("ALLOW_USER_PASSWORD_AUTH"),
				pulumi.String("ALLOW_ADMIN_USER_PASSWORD_AUTH"),
				pulumi.String("ALLOW_REFRESH_TOKEN_AUTH"),
				pulumi.String("ALLOW_USER_SRP_AUTH"),
			},
			GenerateSecret:                  pulumi.Bool(false),
			PreventUserExistenceErrors:      pulumi.String("ENABLED"),
			AllowedOauthFlowsUserPoolClient: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		// Create Admin User Pool (separate pool for admins)
		adminUserPool, err := cognito.NewUserPool(ctx, "askly-admin-pool", &cognito.UserPoolArgs{
			Name: pulumi.String("askly-admins"),
			PasswordPolicy: &cognito.UserPoolPasswordPolicyArgs{
				MinimumLength:    pulumi.Int(12),
				RequireLowercase: pulumi.Bool(true),
				RequireNumbers:   pulumi.Bool(true),
				RequireSymbols:   pulumi.Bool(true),
				RequireUppercase: pulumi.Bool(true),
			},
			AutoVerifiedAttributes: pulumi.StringArray{
				pulumi.String("email"),
			},
			UsernameAttributes: pulumi.StringArray{
				pulumi.String("email"),
			},
			Schemas: cognito.UserPoolSchemaArray{
				&cognito.UserPoolSchemaArgs{
					Name:                   pulumi.String("email"),
					AttributeDataType:      pulumi.String("String"),
					Required:               pulumi.Bool(true),
					Mutable:                pulumi.Bool(true),
					DeveloperOnlyAttribute: pulumi.Bool(false),
				},
				&cognito.UserPoolSchemaArgs{
					Name:                   pulumi.String("role"),
					AttributeDataType:      pulumi.String("String"),
					Required:               pulumi.Bool(false),
					Mutable:                pulumi.Bool(true),
					DeveloperOnlyAttribute: pulumi.Bool(false),
					StringAttributeConstraints: &cognito.UserPoolSchemaStringAttributeConstraintsArgs{
						MinLength: pulumi.String("1"),
						MaxLength: pulumi.String("50"),
					},
				},
			},
			EmailConfiguration: &cognito.UserPoolEmailConfigurationArgs{
				EmailSendingAccount: pulumi.String("COGNITO_DEFAULT"),
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-admin-pool"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Admin Pool Client
		adminPoolClient, err := cognito.NewUserPoolClient(ctx, "askly-admin-pool-client", &cognito.UserPoolClientArgs{
			Name:       pulumi.String("askly-admin-client"),
			UserPoolId: adminUserPool.ID(),
			ExplicitAuthFlows: pulumi.StringArray{
				pulumi.String("ALLOW_USER_PASSWORD_AUTH"),
				pulumi.String("ALLOW_ADMIN_USER_PASSWORD_AUTH"),
				pulumi.String("ALLOW_REFRESH_TOKEN_AUTH"),
				pulumi.String("ALLOW_USER_SRP_AUTH"),
			},
			GenerateSecret:                  pulumi.Bool(false),
			PreventUserExistenceErrors:      pulumi.String("ENABLED"),
			AllowedOauthFlowsUserPoolClient: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		// ===== IAM Role for Lambda Functions =====
		lambdaRole, err := iam.NewRole(ctx, "askly-lambda-role", &iam.RoleArgs{
			Name: pulumi.String("askly-lambda-execution-role"),
			AssumeRolePolicy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Action": "sts:AssumeRole",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					},
					"Effect": "Allow"
				}]
			}`),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-lambda-role"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Attach basic Lambda execution policy
		_, err = iam.NewRolePolicyAttachment(ctx, "lambda-basic-execution", &iam.RolePolicyAttachmentArgs{
			Role:      lambdaRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		})
		if err != nil {
			return err
		}

		// Create inline policy for DynamoDB, Bedrock, and Cognito access
		_, err = iam.NewRolePolicy(ctx, "lambda-dynamodb-bedrock-cognito-policy", &iam.RolePolicyArgs{
			Role: lambdaRole.ID(),
			Policy: pulumi.String(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": [
							"dynamodb:GetItem",
							"dynamodb:PutItem",
							"dynamodb:UpdateItem",
							"dynamodb:DeleteItem",
							"dynamodb:Query",
							"dynamodb:Scan"
						],
						"Resource": [
							"arn:aws:dynamodb:*:*:table/askly-*"
						]
					},
					{
						"Effect": "Allow",
						"Action": [
							"bedrock:InvokeModel",
							"bedrock:ListFoundationModels",
							"bedrock:GetFoundationModel"
						],
						"Resource": "*"
					},
					{
						"Effect": "Allow",
						"Action": [
							"aws-marketplace:ViewSubscriptions",
							"aws-marketplace:Subscribe",
							"aws-marketplace:Unsubscribe"
						],
						"Resource": "*"
					},
					{
						"Effect": "Allow",
						"Action": [
							"cognito-idp:AdminInitiateAuth",
							"cognito-idp:AdminCreateUser",
							"cognito-idp:AdminSetUserPassword",
							"cognito-idp:AdminGetUser",
							"cognito-idp:AdminUpdateUserAttributes",
							"cognito-idp:AdminConfirmSignUp"
						],
						"Resource": "arn:aws:cognito-idp:*:*:userpool/*"
					}
				]
			}`),
		})
		if err != nil {
			return err
		}

		// ===== Lambda Functions =====
		// Lambda environment variables
		lambdaEnv := &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				// DynamoDB Tables
				"DYNAMODB_USERS_TABLE":          usersTable.Name,
				"DYNAMODB_ADMINS_TABLE":         adminsTable.Name,
				"DYNAMODB_CHANNELS_TABLE":       channelsTable.Name,
				"DYNAMODB_TRANSCRIPTIONS_TABLE": transcriptionsTable.Name,
				"DYNAMODB_CHUNKS_TABLE":         chunksTable.Name,
				"DYNAMODB_MESSAGES_TABLE":       messagesTable.Name,

				// AWS Bedrock
				"BEDROCK_EMBEDDING_MODEL": pulumi.String("amazon.titan-embed-text-v2:0"),
				"BEDROCK_CHAT_MODEL":      pulumi.String("anthropic.claude-3-haiku-20240307-v1:0"),

				// Cognito
				"COGNITO_USER_POOL_ID":    userPool.ID(),
				"COGNITO_CLIENT_ID":       userPoolClient.ID(),
				"COGNITO_ADMIN_POOL_ID":   adminUserPool.ID(),
				"COGNITO_ADMIN_CLIENT_ID": adminPoolClient.ID(),

				// Paddle Integration
				"PADDLE_API_KEY":                  pulumi.String(paddleAPIKey),
				"PADDLE_WEBHOOK_SECRET":           pulumi.String(paddleWebhookSecret),
				"PADDLE_ENVIRONMENT":              pulumi.String(paddleEnvironment),
				"PADDLE_PRO_MONTHLY_PRICE_ID":     pulumi.String(paddleProMonthlyPriceID),
				"PADDLE_PRO_ANNUAL_PRICE_ID":      pulumi.String(paddleProAnnualPriceID),
				"PADDLE_PREMIUM_MONTHLY_PRICE_ID": pulumi.String(paddlePremiumMonthlyPriceID),
				"PADDLE_PREMIUM_ANNUAL_PRICE_ID":  pulumi.String(paddlePremiumAnnualPriceID),
			},
		}

		// Lambda: register-user
		registerUserLambda, err := lambda.NewFunction(ctx, "register-user", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-register-user"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/register-user"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-register-user"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: login-user
		loginUserLambda, err := lambda.NewFunction(ctx, "login-user", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-login-user"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/login-user"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-login-user"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: login-admin
		loginAdminLambda, err := lambda.NewFunction(ctx, "login-admin", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-login-admin"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/login-admin"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-login-admin"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: list-channels
		listChannelsLambda, err := lambda.NewFunction(ctx, "list-channels", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-list-channels"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/list-channels"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-list-channels"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: get-channel
		getChannelLambda, err := lambda.NewFunction(ctx, "get-channel", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-get-channel"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/get-channel"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-get-channel"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: upload-transcription
		uploadTranscriptionLambda, err := lambda.NewFunction(ctx, "upload-transcription", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-upload-transcription"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/upload-transcription"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(1024),
			Timeout:     pulumi.Int(60),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-upload-transcription"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: send-message (chat-ask)
		sendMessageLambda, err := lambda.NewFunction(ctx, "send-message", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-send-message"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/send-message"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(1024),
			Timeout:     pulumi.Int(60),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-send-message"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: get-messages (chat-history)
		getMessagesLambda, err := lambda.NewFunction(ctx, "get-messages", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-get-messages"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/get-messages"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-get-messages"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: admin-stats
		adminStatsLambda, err := lambda.NewFunction(ctx, "admin-stats", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-admin-stats"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/admin-stats"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-admin-stats"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: list-transcriptions
		listTranscriptionsLambda, err := lambda.NewFunction(ctx, "list-transcriptions", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-list-transcriptions"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/list-transcriptions"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-list-transcriptions"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: delete-transcription
		deleteTranscriptionLambda, err := lambda.NewFunction(ctx, "delete-transcription", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-delete-transcription"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/delete-transcription"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-delete-transcription"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: get-transcription
		getTranscriptionLambda, err := lambda.NewFunction(ctx, "get-transcription", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-get-transcription"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/get-transcription"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-get-transcription"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: create-channel
		createChannelLambda, err := lambda.NewFunction(ctx, "create-channel", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-create-channel"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/create-channel"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-create-channel"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: create-checkout (Paddle integration)
		createCheckoutLambda, err := lambda.NewFunction(ctx, "create-checkout", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-create-checkout"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/create-checkout"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-create-checkout"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: paddle-webhook (Paddle webhook handler)
		paddleWebhookLambda, err := lambda.NewFunction(ctx, "paddle-webhook", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-paddle-webhook"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/paddle-webhook"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(512),
			Timeout:     pulumi.Int(30),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-paddle-webhook"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: get-subscription (Get user subscription info)
		getSubscriptionLambda, err := lambda.NewFunction(ctx, "get-subscription", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-get-subscription"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/get-subscription"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(256),
			Timeout:     pulumi.Int(10),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-get-subscription"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Lambda: set-channel-selection (Set free user's channel selection)
		setChannelSelectionLambda, err := lambda.NewFunction(ctx, "set-channel-selection", &lambda.FunctionArgs{
			Name:        pulumi.String("askly-set-channel-selection"),
			Runtime:     pulumi.String("provided.al2023"),
			Handler:     pulumi.String("bootstrap"),
			Role:        lambdaRole.Arn,
			Code:        pulumi.NewFileArchive("../backend/lambdas/set-channel-selection"),
			Environment: lambdaEnv,
			MemorySize:  pulumi.Int(256),
			Timeout:     pulumi.Int(10),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-set-channel-selection"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// ===== API Gateway =====
		// Create REST API
		api, err := apigateway.NewRestApi(ctx, "askly-api", &apigateway.RestApiArgs{
			Name:        pulumi.String("askly-api"),
			Description: pulumi.String("Askly POC API Gateway"),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-api"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Create Cognito Authorizer for User Pool
		userAuthorizer, err := apigateway.NewAuthorizer(ctx, "user-cognito-authorizer", &apigateway.AuthorizerArgs{
			Name:         pulumi.String("askly-user-authorizer"),
			RestApi:      api.ID(),
			Type:         pulumi.String("COGNITO_USER_POOLS"),
			ProviderArns: pulumi.StringArray{userPool.Arn},
		})
		if err != nil {
			return err
		}

		// Create Cognito Authorizer for Admin Pool
		adminAuthorizer, err := apigateway.NewAuthorizer(ctx, "admin-cognito-authorizer", &apigateway.AuthorizerArgs{
			Name:         pulumi.String("askly-admin-authorizer"),
			RestApi:      api.ID(),
			Type:         pulumi.String("COGNITO_USER_POOLS"),
			ProviderArns: pulumi.StringArray{adminUserPool.Arn},
		})
		if err != nil {
			return err
		}

		// Get root resource
		rootResourceID := api.RootResourceId

		// Store all integrations for deployment dependencies
		var allIntegrations []pulumi.Resource

		// Helper function to add OPTIONS method with CORS to a resource
		addOptionsMethod := func(resourceID pulumi.IDOutput, resourceName string) error {
			// Create OPTIONS method
			optionsMethod, err := apigateway.NewMethod(ctx, resourceName+"-options-method", &apigateway.MethodArgs{
				RestApi:       api.ID(),
				ResourceId:    resourceID,
				HttpMethod:    pulumi.String("OPTIONS"),
				Authorization: pulumi.String("NONE"),
			})
			if err != nil {
				return err
			}

			// Create MOCK integration for OPTIONS (depends on method)
			optionsIntegration, err := apigateway.NewIntegration(ctx, resourceName+"-options-integration", &apigateway.IntegrationArgs{
				RestApi:    api.ID(),
				ResourceId: resourceID,
				HttpMethod: pulumi.String("OPTIONS"),
				Type:       pulumi.String("MOCK"),
				RequestTemplates: pulumi.StringMap{
					"application/json": pulumi.String(`{"statusCode": 200}`),
				},
			}, pulumi.DependsOn([]pulumi.Resource{optionsMethod}))
			if err != nil {
				return err
			}
			allIntegrations = append(allIntegrations, optionsIntegration)

			// Create OPTIONS method response (depends on method)
			optionsMethodResponse, err := apigateway.NewMethodResponse(ctx, resourceName+"-options-method-response", &apigateway.MethodResponseArgs{
				RestApi:    api.ID(),
				ResourceId: resourceID,
				HttpMethod: pulumi.String("OPTIONS"),
				StatusCode: pulumi.String("200"),
				ResponseParameters: pulumi.BoolMap{
					"method.response.header.Access-Control-Allow-Headers": pulumi.Bool(true),
					"method.response.header.Access-Control-Allow-Methods": pulumi.Bool(true),
					"method.response.header.Access-Control-Allow-Origin":  pulumi.Bool(true),
				},
			}, pulumi.DependsOn([]pulumi.Resource{optionsMethod}))
			if err != nil {
				return err
			}

			// Create OPTIONS integration response (depends on integration AND method response)
			_, err = apigateway.NewIntegrationResponse(ctx, resourceName+"-options-integration-response", &apigateway.IntegrationResponseArgs{
				RestApi:    api.ID(),
				ResourceId: resourceID,
				HttpMethod: pulumi.String("OPTIONS"),
				StatusCode: pulumi.String("200"),
				ResponseParameters: pulumi.StringMap{
					"method.response.header.Access-Control-Allow-Headers": pulumi.String("'Content-Type,Authorization,X-Amz-Date,X-Api-Key,X-Amz-Security-Token'"),
					"method.response.header.Access-Control-Allow-Methods": pulumi.String("'GET,POST,PUT,DELETE,OPTIONS'"),
					"method.response.header.Access-Control-Allow-Origin":  pulumi.String("'*'"),
				},
			}, pulumi.DependsOn([]pulumi.Resource{optionsIntegration, optionsMethodResponse}))
			if err != nil {
				return err
			}

			return nil
		}

		// Helper function to create API Gateway resources and methods
		createAPIRoute := func(pathPart string, parentID pulumi.IDOutput, method string, lambdaFn *lambda.Function, routeName string) error {
			// Create resource
			resource, err := apigateway.NewResource(ctx, routeName+"-resource", &apigateway.ResourceArgs{
				RestApi:  api.ID(),
				ParentId: parentID,
				PathPart: pulumi.String(pathPart),
			})
			if err != nil {
				return err
			}

			// Add OPTIONS method for CORS
			err = addOptionsMethod(resource.ID(), routeName)
			if err != nil {
				return err
			}

			// Create method
			_, err = apigateway.NewMethod(ctx, routeName+"-method", &apigateway.MethodArgs{
				RestApi:       api.ID(),
				ResourceId:    resource.ID(),
				HttpMethod:    pulumi.String(method),
				Authorization: pulumi.String("NONE"),
			})
			if err != nil {
				return err
			}

			// Create integration
			integration, err := apigateway.NewIntegration(ctx, routeName+"-integration", &apigateway.IntegrationArgs{
				RestApi:               api.ID(),
				ResourceId:            resource.ID(),
				HttpMethod:            pulumi.String(method),
				IntegrationHttpMethod: pulumi.String("POST"),
				Type:                  pulumi.String("AWS_PROXY"),
				Uri:                   lambdaFn.InvokeArn,
			})
			if err != nil {
				return err
			}
			allIntegrations = append(allIntegrations, integration)

			// Grant API Gateway permission to invoke Lambda
			_, err = lambda.NewPermission(ctx, routeName+"-permission", &lambda.PermissionArgs{
				Action:    pulumi.String("lambda:InvokeFunction"),
				Function:  lambdaFn.Name,
				Principal: pulumi.String("apigateway.amazonaws.com"),
				SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
			})
			if err != nil {
				return err
			}

			return nil
		}

		// Create /auth resource
		authResource, err := apigateway.NewResource(ctx, "auth-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: rootResourceID,
			PathPart: pulumi.String("auth"),
		})
		if err != nil {
			return err
		}

		// POST /auth/register
		err = createAPIRoute("register", authResource.ID(), "POST", registerUserLambda, "auth-register")
		if err != nil {
			return err
		}

		// POST /auth/login
		err = createAPIRoute("login", authResource.ID(), "POST", loginUserLambda, "auth-login")
		if err != nil {
			return err
		}

		// Create /admin resource
		adminResource, err := apigateway.NewResource(ctx, "admin-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: rootResourceID,
			PathPart: pulumi.String("admin"),
		})
		if err != nil {
			return err
		}

		// POST /admin/login
		err = createAPIRoute("login", adminResource.ID(), "POST", loginAdminLambda, "admin-login")
		if err != nil {
			return err
		}

		// Create /channels resource
		channelsResource, err := apigateway.NewResource(ctx, "channels-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: rootResourceID,
			PathPart: pulumi.String("channels"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /channels
		err = addOptionsMethod(channelsResource.ID(), "channels-list")
		if err != nil {
			return err
		}

		// GET /channels (requires user authentication)
		_, err = apigateway.NewMethod(ctx, "channels-list-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    channelsResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		channelsListIntegration, err := apigateway.NewIntegration(ctx, "channels-list-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            channelsResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   listChannelsLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, channelsListIntegration)

		_, err = lambda.NewPermission(ctx, "channels-list-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  listChannelsLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Create /channels/{id} resource
		channelIDResource, err := apigateway.NewResource(ctx, "channel-id-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: channelsResource.ID(),
			PathPart: pulumi.String("{id}"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /channels/{id}
		err = addOptionsMethod(channelIDResource.ID(), "channel-get")
		if err != nil {
			return err
		}

		// GET /channels/{id} (requires user authentication)
		_, err = apigateway.NewMethod(ctx, "channel-get-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    channelIDResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		channelGetIntegration, err := apigateway.NewIntegration(ctx, "channel-get-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            channelIDResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   getChannelLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, channelGetIntegration)

		_, err = lambda.NewPermission(ctx, "channel-get-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  getChannelLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// POST /channels/{id}/ask
		askResource, err := apigateway.NewResource(ctx, "channel-ask-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: channelIDResource.ID(),
			PathPart: pulumi.String("ask"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /channels/{id}/ask
		err = addOptionsMethod(askResource.ID(), "channel-ask")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "channel-ask-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    askResource.ID(),
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		channelAskIntegration, err := apigateway.NewIntegration(ctx, "channel-ask-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            askResource.ID(),
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   sendMessageLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, channelAskIntegration)

		_, err = lambda.NewPermission(ctx, "channel-ask-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  sendMessageLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// GET /channels/{id}/messages
		messagesResource, err := apigateway.NewResource(ctx, "channel-messages-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: channelIDResource.ID(),
			PathPart: pulumi.String("messages"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /channels/{id}/messages
		err = addOptionsMethod(messagesResource.ID(), "channel-messages")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "channel-messages-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    messagesResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		channelMessagesIntegration, err := apigateway.NewIntegration(ctx, "channel-messages-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            messagesResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   getMessagesLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, channelMessagesIntegration)

		_, err = lambda.NewPermission(ctx, "channel-messages-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  getMessagesLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Create /admin/channels/{id}/transcriptions route
		adminChannelsResource, err := apigateway.NewResource(ctx, "admin-channels-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: adminResource.ID(),
			PathPart: pulumi.String("channels"),
		})
		if err != nil {
			return err
		}

		adminChannelIDResource, err := apigateway.NewResource(ctx, "admin-channel-id-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: adminChannelsResource.ID(),
			PathPart: pulumi.String("{id}"),
		})
		if err != nil {
			return err
		}

		adminTranscriptionsResource, err := apigateway.NewResource(ctx, "admin-transcriptions-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: adminChannelIDResource.ID(),
			PathPart: pulumi.String("transcriptions"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /admin/channels/{id}/transcriptions
		err = addOptionsMethod(adminTranscriptionsResource.ID(), "admin-upload")
		if err != nil {
			return err
		}

		// POST /admin/channels/{id}/transcriptions (requires admin authentication)
		_, err = apigateway.NewMethod(ctx, "admin-upload-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminTranscriptionsResource.ID(),
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminUploadIntegration, err := apigateway.NewIntegration(ctx, "admin-upload-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminTranscriptionsResource.ID(),
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   uploadTranscriptionLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminUploadIntegration)

		_, err = lambda.NewPermission(ctx, "admin-upload-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  uploadTranscriptionLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// GET /admin/stats
		statsResource, err := apigateway.NewResource(ctx, "admin-stats-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: adminResource.ID(),
			PathPart: pulumi.String("stats"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /admin/stats
		err = addOptionsMethod(statsResource.ID(), "admin-stats")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "admin-stats-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    statsResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminStatsIntegration, err := apigateway.NewIntegration(ctx, "admin-stats-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            statsResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   adminStatsLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminStatsIntegration)

		_, err = lambda.NewPermission(ctx, "admin-stats-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  adminStatsLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /admin/channels
		err = addOptionsMethod(adminChannelsResource.ID(), "admin-channels-list")
		if err != nil {
			return err
		}

		// GET /admin/channels
		_, err = apigateway.NewMethod(ctx, "admin-channels-list-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminChannelsResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminChannelsListIntegration, err := apigateway.NewIntegration(ctx, "admin-channels-list-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminChannelsResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   listChannelsLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminChannelsListIntegration)

		_, err = lambda.NewPermission(ctx, "admin-channels-list-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  listChannelsLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// POST /admin/channels
		_, err = apigateway.NewMethod(ctx, "admin-channels-create-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminChannelsResource.ID(),
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminChannelsCreateIntegration, err := apigateway.NewIntegration(ctx, "admin-channels-create-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminChannelsResource.ID(),
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   createChannelLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminChannelsCreateIntegration)

		_, err = lambda.NewPermission(ctx, "admin-channels-create-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  createChannelLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /admin/channels/{id}
		err = addOptionsMethod(adminChannelIDResource.ID(), "admin-channel-get")
		if err != nil {
			return err
		}

		// GET /admin/channels/{id}
		_, err = apigateway.NewMethod(ctx, "admin-channel-get-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminChannelIDResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminChannelGetIntegration, err := apigateway.NewIntegration(ctx, "admin-channel-get-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminChannelIDResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   getChannelLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminChannelGetIntegration)

		_, err = lambda.NewPermission(ctx, "admin-channel-get-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  getChannelLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /admin/channels/{id}/transcriptions (already added above for POST)
		// Skip adding OPTIONS again since it was added for POST method above

		// GET /admin/channels/{id}/transcriptions
		_, err = apigateway.NewMethod(ctx, "admin-list-transcriptions-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminTranscriptionsResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminListTranscriptionsIntegration, err := apigateway.NewIntegration(ctx, "admin-list-transcriptions-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminTranscriptionsResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   listTranscriptionsLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminListTranscriptionsIntegration)

		_, err = lambda.NewPermission(ctx, "admin-list-transcriptions-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  listTranscriptionsLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// DELETE /admin/channels/{id}/transcriptions/{transcriptionId}
		adminTranscriptionIDResource, err := apigateway.NewResource(ctx, "admin-transcription-id-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: adminTranscriptionsResource.ID(),
			PathPart: pulumi.String("{transcriptionId}"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /admin/channels/{id}/transcriptions/{transcriptionId}
		err = addOptionsMethod(adminTranscriptionIDResource.ID(), "admin-delete-transcription")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "admin-delete-transcription-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminTranscriptionIDResource.ID(),
			HttpMethod:    pulumi.String("DELETE"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminDeleteTranscriptionIntegration, err := apigateway.NewIntegration(ctx, "admin-delete-transcription-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminTranscriptionIDResource.ID(),
			HttpMethod:            pulumi.String("DELETE"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   deleteTranscriptionLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminDeleteTranscriptionIntegration)

		_, err = lambda.NewPermission(ctx, "admin-delete-transcription-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  deleteTranscriptionLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// GET /admin/channels/{id}/transcriptions/{transcriptionId}
		_, err = apigateway.NewMethod(ctx, "admin-get-transcription-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    adminTranscriptionIDResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  adminAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		adminGetTranscriptionIntegration, err := apigateway.NewIntegration(ctx, "admin-get-transcription-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            adminTranscriptionIDResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   getTranscriptionLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, adminGetTranscriptionIntegration)

		_, err = lambda.NewPermission(ctx, "admin-get-transcription-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  getTranscriptionLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// ===== Paddle Integration Routes =====
		// POST /checkout - Create Paddle checkout session (requires user auth)
		checkoutResource, err := apigateway.NewResource(ctx, "checkout-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: rootResourceID,
			PathPart: pulumi.String("checkout"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /checkout
		err = addOptionsMethod(checkoutResource.ID(), "checkout")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "checkout-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    checkoutResource.ID(),
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		checkoutIntegration, err := apigateway.NewIntegration(ctx, "checkout-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            checkoutResource.ID(),
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   createCheckoutLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, checkoutIntegration)

		_, err = lambda.NewPermission(ctx, "checkout-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  createCheckoutLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Create /webhook resource
		webhookResource, err := apigateway.NewResource(ctx, "webhook-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: rootResourceID,
			PathPart: pulumi.String("webhook"),
		})
		if err != nil {
			return err
		}

		// POST /webhook/paddle - Paddle webhook handler (NO AUTH - public endpoint)
		paddleWebhookResource, err := apigateway.NewResource(ctx, "webhook-paddle-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: webhookResource.ID(),
			PathPart: pulumi.String("paddle"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /webhook/paddle
		err = addOptionsMethod(paddleWebhookResource.ID(), "webhook-paddle")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "webhook-paddle-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    paddleWebhookResource.ID(),
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("NONE"), // NO AUTH - Paddle will call this
		})
		if err != nil {
			return err
		}

		webhookPaddleIntegration, err := apigateway.NewIntegration(ctx, "webhook-paddle-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            paddleWebhookResource.ID(),
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   paddleWebhookLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, webhookPaddleIntegration)

		_, err = lambda.NewPermission(ctx, "webhook-paddle-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  paddleWebhookLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// ===== User Subscription Routes =====
		// Create /user resource
		userResource, err := apigateway.NewResource(ctx, "user-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: rootResourceID,
			PathPart: pulumi.String("user"),
		})
		if err != nil {
			return err
		}

		// GET /user/subscription - Get user subscription info (requires user auth)
		userSubscriptionResource, err := apigateway.NewResource(ctx, "user-subscription-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: userResource.ID(),
			PathPart: pulumi.String("subscription"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /user/subscription
		err = addOptionsMethod(userSubscriptionResource.ID(), "user-subscription")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "user-subscription-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    userSubscriptionResource.ID(),
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		userSubscriptionIntegration, err := apigateway.NewIntegration(ctx, "user-subscription-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            userSubscriptionResource.ID(),
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   getSubscriptionLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, userSubscriptionIntegration)

		_, err = lambda.NewPermission(ctx, "user-subscription-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  getSubscriptionLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// POST /user/channel-selection - Set free user's channel selection (requires user auth)
		userChannelSelectionResource, err := apigateway.NewResource(ctx, "user-channel-selection-resource", &apigateway.ResourceArgs{
			RestApi:  api.ID(),
			ParentId: userResource.ID(),
			PathPart: pulumi.String("channel-selection"),
		})
		if err != nil {
			return err
		}

		// Add OPTIONS to /user/channel-selection
		err = addOptionsMethod(userChannelSelectionResource.ID(), "user-channel-selection")
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "user-channel-selection-method", &apigateway.MethodArgs{
			RestApi:       api.ID(),
			ResourceId:    userChannelSelectionResource.ID(),
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("COGNITO_USER_POOLS"),
			AuthorizerId:  userAuthorizer.ID(),
		})
		if err != nil {
			return err
		}

		userChannelSelectionIntegration, err := apigateway.NewIntegration(ctx, "user-channel-selection-integration", &apigateway.IntegrationArgs{
			RestApi:               api.ID(),
			ResourceId:            userChannelSelectionResource.ID(),
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   setChannelSelectionLambda.InvokeArn,
		})
		if err != nil {
			return err
		}
		allIntegrations = append(allIntegrations, userChannelSelectionIntegration)

		_, err = lambda.NewPermission(ctx, "user-channel-selection-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  setChannelSelectionLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

		// Deploy API Gateway
		deployment, err := apigateway.NewDeployment(ctx, "api-deployment", &apigateway.DeploymentArgs{
			RestApi: api.ID(),
			// Force new deployment on every update
			Triggers: pulumi.StringMap{
				"redeployment": pulumi.Sprintf("%v", "7"), // Incremented for user subscription routes
			},
		}, pulumi.DependsOn(allIntegrations))
		if err != nil {
			return err
		}

		// Create stage
		stage, err := apigateway.NewStage(ctx, "api-stage", &apigateway.StageArgs{
			RestApi:    api.ID(),
			Deployment: deployment.ID(),
			StageName:  pulumi.String("prod"),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-api-prod"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// ===== Exports =====
		ctx.Export("usersTableName", usersTable.Name)
		ctx.Export("adminsTableName", adminsTable.Name)
		ctx.Export("channelsTableName", channelsTable.Name)
		ctx.Export("transcriptionsTableName", transcriptionsTable.Name)
		ctx.Export("chunksTableName", chunksTable.Name)
		ctx.Export("messagesTableName", messagesTable.Name)
		ctx.Export("lambdaRoleArn", lambdaRole.Arn)
		ctx.Export("apiGatewayURL", pulumi.Sprintf("https://%s.execute-api.us-east-1.amazonaws.com/%s", api.ID(), stage.StageName))
		ctx.Export("apiGatewayID", api.ID())

		// Cognito exports
		ctx.Export("cognitoUserPoolId", userPool.ID())
		ctx.Export("cognitoUserPoolArn", userPool.Arn)
		ctx.Export("cognitoClientId", userPoolClient.ID())
		ctx.Export("cognitoAdminPoolId", adminUserPool.ID())
		ctx.Export("cognitoAdminPoolArn", adminUserPool.Arn)
		ctx.Export("cognitoAdminClientId", adminPoolClient.ID())

		return nil
	})
}
