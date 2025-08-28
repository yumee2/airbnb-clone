package entity

import "time"

type Profile struct {
	ID          string    `gorm:"type:uuid;primaryKey"`
	PhoneNumber string    `gorm:"size:16;not null;unique"`
	Name        string    `gorm:"size:255;not null"`
	Surname     string    `gorm:"size:255"`
	DateOfBirth time.Time `gorm:"type:date;not null"`
	ImagePath   string    `gorm:"size:500"`
}

type CreateProfileRequest struct {
	PhoneNumber string    `json:"phone_number" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Surname     string    `json:"surname"`
	DateOfBirth time.Time `json:"date_of_birth" binding:"required"`
	Image       []byte    `json:"-"`
}

type ProfileResponse struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	DateOfBirth time.Time `json:"date_of_birth"`
	ImageURL    string    `json:"image_url"`
}
