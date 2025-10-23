package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cognito"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
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
			},
			GlobalSecondaryIndexes: dynamodb.TableGlobalSecondaryIndexArray{
				&dynamodb.TableGlobalSecondaryIndexArgs{
					Name:           pulumi.String("email-index"),
					HashKey:        pulumi.String("email"),
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
				"DYNAMODB_USERS_TABLE":          usersTable.Name,
				"DYNAMODB_ADMINS_TABLE":         adminsTable.Name,
				"DYNAMODB_CHANNELS_TABLE":       channelsTable.Name,
				"DYNAMODB_TRANSCRIPTIONS_TABLE": transcriptionsTable.Name,
				"DYNAMODB_CHUNKS_TABLE":         chunksTable.Name,
				"DYNAMODB_MESSAGES_TABLE":       messagesTable.Name,
				"BEDROCK_EMBEDDING_MODEL":       pulumi.String("amazon.titan-embed-text-v2:0"),
				"BEDROCK_CHAT_MODEL":            pulumi.String("anthropic.claude-3-haiku-20240307-v1:0"),
				"COGNITO_USER_POOL_ID":          userPool.ID(),
				"COGNITO_CLIENT_ID":             userPoolClient.ID(),
				"COGNITO_ADMIN_POOL_ID":         adminUserPool.ID(),
				"COGNITO_ADMIN_CLIENT_ID":       adminPoolClient.ID(),
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
			_, err = apigateway.NewIntegration(ctx, routeName+"-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "channels-list-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "channel-get-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "channel-ask-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "channel-messages-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "admin-upload-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "admin-stats-integration", &apigateway.IntegrationArgs{
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

		_, err = lambda.NewPermission(ctx, "admin-stats-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  adminStatsLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
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

		_, err = apigateway.NewIntegration(ctx, "admin-channels-list-integration", &apigateway.IntegrationArgs{
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

		_, err = lambda.NewPermission(ctx, "admin-channels-list-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  listChannelsLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
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

		_, err = apigateway.NewIntegration(ctx, "admin-channel-get-integration", &apigateway.IntegrationArgs{
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

		_, err = lambda.NewPermission(ctx, "admin-channel-get-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  getChannelLambda.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("%s/*/*", api.ExecutionArn),
		})
		if err != nil {
			return err
		}

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

		_, err = apigateway.NewIntegration(ctx, "admin-list-transcriptions-integration", &apigateway.IntegrationArgs{
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

		_, err = apigateway.NewIntegration(ctx, "admin-delete-transcription-integration", &apigateway.IntegrationArgs{
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

		_, err = lambda.NewPermission(ctx, "admin-delete-transcription-permission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  deleteTranscriptionLambda.Name,
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
				"redeployment": pulumi.Sprintf("%v", "2"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{
			registerUserLambda,
			loginUserLambda,
			loginAdminLambda,
			listChannelsLambda,
			getChannelLambda,
			uploadTranscriptionLambda,
			sendMessageLambda,
			getMessagesLambda,
			adminStatsLambda,
			listTranscriptionsLambda,
			deleteTranscriptionLambda,
		}))
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
