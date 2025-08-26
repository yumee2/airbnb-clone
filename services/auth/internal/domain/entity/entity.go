package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type UserCredentials struct {
	ID       string `gorm:"type:uuid;primaryKey"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"size:255;not null"`
}

func (u *UserCredentials) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

type RefreshToken struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex"` // Hashed token
	UserID    string    `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

func (r *RefreshToken) HashToken(token string) {
	hash := sha256.Sum256([]byte(token))
	r.TokenHash = hex.EncodeToString(hash[:])
}

func (r *RefreshToken) IsValid() bool {
	return time.Now().Before(r.ExpiresAt)
}
