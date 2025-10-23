package models

import "time"

// User represents a user in the system
type User struct {
	PK           string    `dynamodbav:"PK" json:"id"`
	SK           string    `dynamodbav:"SK" json:"-"`
	Email        string    `dynamodbav:"email" json:"email"`
	PasswordHash string    `dynamodbav:"password_hash" json:"-"`
	CreatedAt    time.Time `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt    time.Time `dynamodbav:"updated_at" json:"updated_at"`
}

// Admin represents an admin user
type Admin struct {
	PK           string    `dynamodbav:"PK" json:"id"`
	SK           string    `dynamodbav:"SK" json:"-"`
	Email        string    `dynamodbav:"email" json:"email"`
	PasswordHash string    `dynamodbav:"password_hash" json:"-"`
	CreatedAt    time.Time `dynamodbav:"created_at" json:"created_at"`
}

// NewUser creates a new User with generated ID
func NewUser(email, passwordHash string) *User {
	now := time.Now()
	return &User{
		PK:           GenerateID(),
		SK:           "PROFILE",
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewAdmin creates a new Admin with generated ID
func NewAdmin(email, passwordHash string) *Admin {
	return &Admin{
		PK:           GenerateID(),
		SK:           "PROFILE",
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
}

// GetID returns the ID (PK is now the ID directly)
func (u *User) GetID() string {
	return u.PK
}

// GetID returns the ID (PK is now the ID directly)
func (a *Admin) GetID() string {
	return a.PK
}
