package model

import "time"

type TokenType string

const (
	TokenTypeEmailVerify   TokenType = "email_verify"
	TokenTypePasswordReset TokenType = "password_reset"
)

type OneTimeToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `gorm:"index;not null" json:"user_id"`
	TokenHash string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	Type      TokenType  `gorm:"size:32;index;not null" json:"type"`
	ExpiresAt time.Time  `gorm:"index;not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
