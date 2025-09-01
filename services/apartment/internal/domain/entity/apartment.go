package entity

import "time"

type Apartment struct {
	ID            string  `gorm:"primaryKey"`
	HostID        string  `gorm:"not null"`
	Title         string  `gorm:"size:255;not null"`
	Description   string  `gorm:"type:text"`
	PricePerNight float64 `gorm:"not null"`

	HouseNumber int    `gorm:"not null"`
	Street      string `gorm:"size:255;not null"`
	City        string `gorm:"size:100;not null"`
	State       string `gorm:"size:100"`
	Country     string `gorm:"size:100;not null"`
	PostalCode  string `gorm:"size:20"`

	Latitude  float64
	Longitude float64

	Wifi         bool `gorm:"default:false"`
	Parking      bool `gorm:"default:false"`
	AirCondition bool `gorm:"default:false"`
	Kitchen      bool `gorm:"default:false"`
	PetFriendly  bool `gorm:"default:false"`

	MaxGuests     int     `gorm:"not null"`
	BedroomNumber int     `gorm:"not null"`
	Images        []Image `gorm:"foreignKey:ApartmentID;constraint:OnDelete:CASCADE;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
