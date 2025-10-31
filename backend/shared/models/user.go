package models

import "time"

// User represents a user in the system
type User struct {
	PK           string `dynamodbav:"PK" json:"id"`
	SK           string `dynamodbav:"SK" json:"-"`
	Email        string `dynamodbav:"email" json:"email"`
	PasswordHash string `dynamodbav:"password_hash" json:"-"`

	// Subscription fields
	SubscriptionTier     string `dynamodbav:"subscription_tier" json:"subscription_tier"`                               // "free", "pro", "premium"
	PaddleCustomerID     string `dynamodbav:"paddle_customer_id,omitempty" json:"paddle_customer_id,omitempty"`         // Paddle customer ID
	PaddleSubscriptionID string `dynamodbav:"paddle_subscription_id,omitempty" json:"paddle_subscription_id,omitempty"` // Paddle subscription ID
	SubscriptionStatus   string `dynamodbav:"subscription_status" json:"subscription_status"`                           // "active", "canceled", "past_due"

	// Usage tracking
	QuestionsUsedThisMonth int    `dynamodbav:"questions_used_this_month" json:"questions_used_this_month"`
	QuestionsLimit         int    `dynamodbav:"questions_limit" json:"questions_limit"`
	ChannelsLimit          int    `dynamodbav:"channels_limit" json:"channels_limit"`                               // -1 = unlimited
	SelectedChannelID      string `dynamodbav:"selected_channel_id,omitempty" json:"selected_channel_id,omitempty"` // For free users

	// Billing period tracking
	CurrentPeriodStart *time.Time `dynamodbav:"current_period_start,omitempty" json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `dynamodbav:"current_period_end,omitempty" json:"current_period_end,omitempty"`

	CreatedAt time.Time `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at" json:"updated_at"`
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
	nextMonth := now.AddDate(0, 1, 0)
	return &User{
		PK:                     GenerateID(),
		SK:                     "PROFILE",
		Email:                  email,
		PasswordHash:           passwordHash,
		SubscriptionTier:       "free", // Default to free tier
		SubscriptionStatus:     "active",
		QuestionsUsedThisMonth: 0,
		QuestionsLimit:         5, // Free tier: 5 questions/month
		ChannelsLimit:          1, // Free tier: 1 channel only
		CurrentPeriodStart:     &now,
		CurrentPeriodEnd:       &nextMonth,
		CreatedAt:              now,
		UpdatedAt:              now,
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

// CanAskQuestion checks if the user has questions remaining for this period
func (u *User) CanAskQuestion() bool {
	return u.QuestionsUsedThisMonth < u.QuestionsLimit
}

// RemainingQuestions returns how many questions the user has left this period
func (u *User) RemainingQuestions() int {
	remaining := u.QuestionsLimit - u.QuestionsUsedThisMonth
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CanAccessChannel checks if the user can access a specific channel based on their tier
func (u *User) CanAccessChannel(channelID string) bool {
	// Unlimited access (premium tier)
	if u.ChannelsLimit == -1 {
		return true
	}

	// Free tier: can only access selected channel
	if u.SubscriptionTier == "free" {
		// If no channel selected yet, allow access (will be set on first access)
		if u.SelectedChannelID == "" {
			return true
		}
		return u.SelectedChannelID == channelID
	}

	// Pro and Premium: access to all available channels
	return true
}

// NeedsReset checks if the user's billing period has ended and usage should be reset
func (u *User) NeedsReset() bool {
	if u.CurrentPeriodEnd == nil {
		return true
	}
	return time.Now().After(*u.CurrentPeriodEnd)
}
