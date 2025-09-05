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

	MaxGuests     int       `gorm:"not null"`
	BedroomNumber int       `gorm:"not null"`
	Images        []Image   `gorm:"foreignKey:ApartmentID;constraint:OnDelete:CASCADE;"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoCreateTime"`
}

type CreateApartmentRequest struct {
	Title         string  `form:"title" binding:"required"`
	Description   string  `form:"description"`
	PricePerNight float64 `form:"price_per_night" binding:"required"`

	HouseNumber int    `form:"house_number" binding:"required"`
	Street      string `form:"street" binding:"required"`
	City        string `form:"city" binding:"required"`
	State       string `form:"state"`
	Country     string `form:"country" binding:"required"`
	PostalCode  string `form:"postal_code"`

	Latitude  float64 `form:"latitude"`
	Longitude float64 `form:"longitude"`

	Wifi         bool `form:"wifi"`
	Parking      bool `form:"parking"`
	AirCondition bool `form:"air_condition"`
	Kitchen      bool `form:"kitchen"`
	PetFriendly  bool `form:"pet_friendly"`

	MaxGuests     int `form:"max_guests" binding:"required"`
	BedroomNumber int `form:"bedroom_number" binding:"required"`
}

type UpdateApartmentRequest struct {
	Title         *string  `json:"title,omitempty"`
	Description   *string  `json:"description,omitempty"`
	PricePerNight *float64 `json:"price_per_night,omitempty"`

	HouseNumber *int    `json:"house_number,omitempty"`
	Street      *string `json:"street,omitempty"`
	City        *string `json:"city,omitempty"`
	State       *string `json:"state,omitempty"`
	Country     *string `json:"country,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`

	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	Wifi         *bool `json:"wifi,omitempty"`
	Parking      *bool `json:"parking,omitempty"`
	AirCondition *bool `json:"air_condition,omitempty"`
	Kitchen      *bool `json:"kitchen,omitempty"`
	PetFriendly  *bool `json:"pet_friendly,omitempty"`

	MaxGuests     *int `json:"max_guests,omitempty"`
	BedroomNumber *int `json:"bedroom_number,omitempty"`
}

type ApartmentResponse struct {
	ID            string  `json:"id"`
	HostID        string  `json:"host_id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	PricePerNight float64 `json:"price_per_night"`

	HouseNumber int    `json:"house_number"`
	Street      string `json:"street"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Wifi         bool `json:"wifi"`
	Parking      bool `json:"parking"`
	AirCondition bool `json:"air_condition"`
	Kitchen      bool `json:"kitchen"`
	PetFriendly  bool `json:"pet_friendly"`

	MaxGuests     int `json:"max_guests"`
	BedroomNumber int `json:"bedroom_number"`

	Images    []ImageResponse `json:"images"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
