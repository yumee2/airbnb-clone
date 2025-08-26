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
