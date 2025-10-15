package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudfront"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
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

		// ===== S3 Bucket for Frontend =====
		frontendBucket, err := s3.NewBucket(ctx, "askly-frontend", &s3.BucketArgs{
			Bucket: pulumi.String("askly-frontend-poc"),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-frontend"),
				"Environment": pulumi.String("poc"),
			},
		})
		if err != nil {
			return err
		}

		// Configure bucket for website hosting
		_, err = s3.NewBucketWebsiteConfigurationV2(ctx, "askly-frontend-website", &s3.BucketWebsiteConfigurationV2Args{
			Bucket: frontendBucket.ID(),
			IndexDocument: &s3.BucketWebsiteConfigurationV2IndexDocumentArgs{
				Suffix: pulumi.String("index.html"),
			},
			ErrorDocument: &s3.BucketWebsiteConfigurationV2ErrorDocumentArgs{
				Key: pulumi.String("index.html"), // SPA fallback
			},
		})
		if err != nil {
			return err
		}

		// Make bucket public for CloudFront
		_, err = s3.NewBucketPublicAccessBlock(ctx, "askly-frontend-public-access", &s3.BucketPublicAccessBlockArgs{
			Bucket:                frontendBucket.ID(),
			BlockPublicAcls:       pulumi.Bool(false),
			BlockPublicPolicy:     pulumi.Bool(false),
			IgnorePublicAcls:      pulumi.Bool(false),
			RestrictPublicBuckets: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		// Bucket policy for CloudFront access
		_, err = s3.NewBucketPolicy(ctx, "askly-frontend-policy", &s3.BucketPolicyArgs{
			Bucket: frontendBucket.ID(),
			Policy: pulumi.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [{
					"Sid": "PublicReadGetObject",
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::%s/*"
				}]
			}`, frontendBucket.ID()),
		})
		if err != nil {
			return err
		}

		// ===== CloudFront Distribution =====
		cdn, err := cloudfront.NewDistribution(ctx, "askly-cdn", &cloudfront.DistributionArgs{
			Origins: cloudfront.DistributionOriginArray{
				&cloudfront.DistributionOriginArgs{
					DomainName: frontendBucket.BucketRegionalDomainName,
					OriginId:   pulumi.String("S3-askly-frontend"),
				},
			},
			Enabled:           pulumi.Bool(true),
			IsIpv6Enabled:     pulumi.Bool(true),
			Comment:           pulumi.String("Askly POC CDN"),
			DefaultRootObject: pulumi.String("index.html"),
			DefaultCacheBehavior: &cloudfront.DistributionDefaultCacheBehaviorArgs{
				AllowedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
					pulumi.String("OPTIONS"),
				},
				CachedMethods: pulumi.StringArray{
					pulumi.String("GET"),
					pulumi.String("HEAD"),
				},
				TargetOriginId:       pulumi.String("S3-askly-frontend"),
				ViewerProtocolPolicy: pulumi.String("redirect-to-https"),
				ForwardedValues: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesArgs{
					QueryString: pulumi.Bool(false),
					Cookies: &cloudfront.DistributionDefaultCacheBehaviorForwardedValuesCookiesArgs{
						Forward: pulumi.String("none"),
					},
				},
				MinTtl:     pulumi.Int(0),
				DefaultTtl: pulumi.Int(3600),
				MaxTtl:     pulumi.Int(86400),
			},
			// Custom error response for SPA routing
			CustomErrorResponses: cloudfront.DistributionCustomErrorResponseArray{
				&cloudfront.DistributionCustomErrorResponseArgs{
					ErrorCode:        pulumi.Int(404),
					ResponseCode:     pulumi.Int(200),
					ResponsePagePath: pulumi.String("/index.html"),
				},
			},
			PriceClass: pulumi.String("PriceClass_100"), // Use only North America and Europe
			Restrictions: &cloudfront.DistributionRestrictionsArgs{
				GeoRestriction: &cloudfront.DistributionRestrictionsGeoRestrictionArgs{
					RestrictionType: pulumi.String("none"),
				},
			},
			ViewerCertificate: &cloudfront.DistributionViewerCertificateArgs{
				CloudfrontDefaultCertificate: pulumi.Bool(true),
			},
			Tags: pulumi.StringMap{
				"Name":        pulumi.String("askly-cdn"),
				"Environment": pulumi.String("poc"),
			},
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

		// Create inline policy for DynamoDB and Bedrock access
		_, err = iam.NewRolePolicy(ctx, "lambda-dynamodb-bedrock-policy", &iam.RolePolicyArgs{
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
							"bedrock:InvokeModel"
						],
						"Resource": "*"
					}
				]
			}`),
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
		ctx.Export("frontendBucketName", frontendBucket.ID())
		ctx.Export("cloudFrontUrl", cdn.DomainName)
		ctx.Export("cloudFrontDistributionId", cdn.ID())
		ctx.Export("lambdaRoleArn", lambdaRole.Arn)

		return nil
	})
}
