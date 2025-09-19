package repository

import (
	"airbnb-clone/apt/internal/config"
	"airbnb-clone/apt/internal/domain/entity"
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ApartmentRepository interface {
	CreateNewApartment(apartment *entity.Apartment) error
	GetApartment(id string) (*entity.Apartment, error)
	DeleteApartmentByID(id string) error
	UpdateApartmentFields(id string, updates map[string]interface{}) error
	AddImages(apartmentID string, images []entity.Image) error
	GetApartmentImages(apartmentID string) ([]entity.Image, error)
	DeleteApartmentImages(apartmentID string) error
}

type storage struct {
	db *gorm.DB
}

func New(cfg *config.Config) (ApartmentRepository, error) {
	const fn = "adapters.repository.New"
	dsn := fmt.Sprintf("host=%s user=%s "+
		"password=%s dbname=%s port=%d sslmode=disable",
		cfg.PostgresConnect.Host, cfg.PostgresConnect.User, cfg.PostgresConnect.Password, cfg.PostgresConnect.DatabaseName, cfg.PostgresConnect.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	err = db.AutoMigrate(&entity.Apartment{}, &entity.Image{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &storage{db: db}, nil
}

func (s *storage) CreateNewApartment(apartment *entity.Apartment) error {
	const fn = "adapters.repository.CreateNewApartment"
	result := s.db.Create(&apartment)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", fn, result.Error)
	}

	return nil
}

func (s *storage) GetApartment(id string) (*entity.Apartment, error) {
	const fn = "adapters.repository.GetApartment"
	var apt entity.Apartment

	result := s.db.Preload("Images").First(&apt, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.Apartment{}, ErrAptNotFound
		}

		return &entity.Apartment{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return &apt, nil
}

func (s *storage) DeleteApartmentByID(id string) error {
	const fn = "adapters.repository.DeleteApartmentByID"

	result := s.db.Delete(&entity.Apartment{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrAptNotFound
	}
	return nil
}

func (s *storage) UpdateApartmentFields(id string, updates map[string]interface{}) error {
	const fn = "adapters.repository.UpdateApartmentFields"

	result := s.db.Model(&entity.Apartment{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrAptNotFound
	}

	return nil
}

func (s *storage) AddImages(apartmentID string, images []entity.Image) error {
	for i := range images {
		images[i].ApartmentID = apartmentID
	}
	return s.db.Create(&images).Error
}

func (s *storage) GetApartmentImages(apartmentID string) ([]entity.Image, error) {
	var images []entity.Image
	err := s.db.Where("apartment_id = ?", apartmentID).Find(&images).Error
	return images, err
}

func (s *storage) DeleteApartmentImages(apartmentID string) error {
	const fn = "adapters.repository.DeleteApartmentImages"

	result := s.db.Delete(&entity.Image{}, "apartment_id = ?", apartmentID)
	if result.Error != nil {
		return fmt.Errorf("%s: database error: %w", fn, result.Error)
	}
	return nil
}
