package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBRepository provides DynamoDB operations
type DynamoDBRepository struct {
	client *dynamodb.Client
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(ctx context.Context) (*DynamoDBRepository, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &DynamoDBRepository{
		client: dynamodb.NewFromConfig(cfg),
	}, nil
}

// NewDynamoDBRepositoryWithClient creates a repository with an existing client
func NewDynamoDBRepositoryWithClient(client *dynamodb.Client) *DynamoDBRepository {
	return &DynamoDBRepository{
		client: client,
	}
}

// PutItem puts an item into DynamoDB
func (r *DynamoDBRepository) PutItem(ctx context.Context, tableName string, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// GetItem gets an item from DynamoDB by PK and SK
func (r *DynamoDBRepository) GetItem(ctx context.Context, tableName, pk, sk string, result interface{}) error {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	}

	output, err := r.client.GetItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	if output.Item == nil {
		return fmt.Errorf("item not found")
	}

	if err := attributevalue.UnmarshalMap(output.Item, result); err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return nil
}

// QueryByPK queries items by PK
func (r *DynamoDBRepository) QueryByPK(ctx context.Context, tableName, pk string, results interface{}) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
		},
	}

	output, err := r.client.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to query by PK: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(output.Items, results); err != nil {
		return fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return nil
}

// QueryByGSI queries items using a Global Secondary Index
func (r *DynamoDBRepository) QueryByGSI(ctx context.Context, tableName, indexName, keyName, keyValue string, results interface{}) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :key", keyName)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":key": &types.AttributeValueMemberS{Value: keyValue},
		},
	}

	output, err := r.client.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to query GSI: %w", err)
	}

	if len(output.Items) == 0 {
		return fmt.Errorf("no items found")
	}

	if err := attributevalue.UnmarshalListOfMaps(output.Items, results); err != nil {
		return fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return nil
}

// QueryByGSIRaw queries items using a Global Secondary Index and returns raw DynamoDB items
func (r *DynamoDBRepository) QueryByGSIRaw(ctx context.Context, tableName, indexName, keyName, keyValue string) ([]map[string]types.AttributeValue, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String(indexName),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :key", keyName)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":key": &types.AttributeValueMemberS{Value: keyValue},
		},
	}

	output, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query GSI: %w", err)
	}

	return output.Items, nil
}

// Scan scans all items in a table (use sparingly)
func (r *DynamoDBRepository) Scan(ctx context.Context, tableName string, results interface{}) error {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	output, err := r.client.Scan(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to scan table: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(output.Items, results); err != nil {
		return fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return nil
}

// DeleteItem deletes an item from DynamoDB
func (r *DynamoDBRepository) DeleteItem(ctx context.Context, tableName, pk, sk string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	}

	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// UpdateItem updates specific attributes of an item
func (r *DynamoDBRepository) UpdateItem(ctx context.Context, tableName, pk, sk string, updates map[string]interface{}) error {
	// Build update expression
	updateExpr := "SET "
	exprAttrValues := make(map[string]types.AttributeValue)
	exprAttrNames := make(map[string]string)

	i := 0
	for key, value := range updates {
		if i > 0 {
			updateExpr += ", "
		}
		placeholder := fmt.Sprintf(":val%d", i)
		attrName := fmt.Sprintf("#attr%d", i)
		updateExpr += fmt.Sprintf("%s = %s", attrName, placeholder)

		av, err := attributevalue.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal update value: %w", err)
		}
		exprAttrValues[placeholder] = av
		exprAttrNames[attrName] = key
		i++
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
	}

	_, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}
