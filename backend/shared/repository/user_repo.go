package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

const (
	usersTableEnv  = "DYNAMODB_USERS_TABLE"
	adminsTableEnv = "DYNAMODB_ADMINS_TABLE"
)

// UserRepository provides user-specific operations
type UserRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewUserRepository creates a new user repository
func NewUserRepository(ctx context.Context) (*UserRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, err
	}

	tableName := os.Getenv(usersTableEnv)
	if tableName == "" {
		tableName = "askly-users" // Default fallback
	}

	return &UserRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user interface{}) error {
	return r.db.PutItem(ctx, r.tableName, user)
}

// GetUserByID gets a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID string, result interface{}) error {
	return r.db.GetItem(ctx, r.tableName, userID, "PROFILE", result)
}

// GetUserByEmail gets a user by email using GSI
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string, result interface{}) error {
	items, err := r.db.QueryByGSIRaw(ctx, r.tableName, "email-index", "email", email)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("user not found")
	}

	// Unmarshal the first user into the result
	if err := attributevalue.UnmarshalMap(items[0], result); err != nil {
		return fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return nil
}

// UpdateUser updates user attributes
func (r *UserRepository) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error {
	return r.db.UpdateItem(ctx, r.tableName, userID, "PROFILE", updates)
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	return r.db.DeleteItem(ctx, r.tableName, userID, "PROFILE")
}

// AdminRepository provides admin-specific operations
type AdminRepository struct {
	db        *DynamoDBRepository
	tableName string
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(ctx context.Context) (*AdminRepository, error) {
	db, err := NewDynamoDBRepository(ctx)
	if err != nil {
		return nil, err
	}

	tableName := os.Getenv(adminsTableEnv)
	if tableName == "" {
		tableName = "askly-admins" // Default fallback
	}

	return &AdminRepository{
		db:        db,
		tableName: tableName,
	}, nil
}

// CreateAdmin creates a new admin
func (r *AdminRepository) CreateAdmin(ctx context.Context, admin interface{}) error {
	return r.db.PutItem(ctx, r.tableName, admin)
}

// GetAdminByID gets an admin by ID
func (r *AdminRepository) GetAdminByID(ctx context.Context, adminID string, result interface{}) error {
	return r.db.GetItem(ctx, r.tableName, adminID, "PROFILE", result)
}

// GetAdminByEmail gets an admin by email using GSI
func (r *AdminRepository) GetAdminByEmail(ctx context.Context, email string, result interface{}) error {
	items, err := r.db.QueryByGSIRaw(ctx, r.tableName, "email-index", "email", email)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("admin not found")
	}

	// Unmarshal the first admin into the result
	if err := attributevalue.UnmarshalMap(items[0], result); err != nil {
		return fmt.Errorf("failed to unmarshal admin: %w", err)
	}

	return nil
}
