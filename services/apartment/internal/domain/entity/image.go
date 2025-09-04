package entity

type Image struct {
	ID          string `gorm:"primaryKey"`
	ApartmentID string `gorm:"index"`
	Path        string `gorm:"size:255;not null"`
	IsCover     bool   `gorm:"default:false"`
}

type ImageResponse struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	IsCover bool   `json:"is_cover"`
}
