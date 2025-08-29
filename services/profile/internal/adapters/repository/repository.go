package repository

import (
	"errors"
	"fmt"

	"airbnb.com/services/profile/internal/config"
	"airbnb.com/services/profile/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ProfileRepository interface {
	CreateNewProfile(profile *entity.Profile) error
	GetUserById(userId string) (*entity.Profile, error)
	GetMe(userId string) (*entity.Profile, error)
	DeleteProfileByID(id string) error
	UpdateProfileFields(id string, updates map[string]interface{}) error
}

type storage struct {
	db *gorm.DB
}

func New(cfg *config.Config) (ProfileRepository, error) {
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

	err = db.AutoMigrate(&entity.Profile{}) // domain models
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &storage{db: db}, nil
}

func (s *storage) CreateNewProfile(profile *entity.Profile) error {
	const fn = "adapters.repository.CreateNewProfile"

	result := s.db.Create(&profile)
	var pgErr *pgconn.PgError

	if result.Error != nil {
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return ErrPhoneNumberExist
		}
		return fmt.Errorf("%s: %w", fn, result.Error)
	}

	return nil
}

func (s *storage) GetUserById(userId string) (*entity.Profile, error) {
	const fn = "adapters.repository.GetUserById"
	var profile *entity.Profile

	result := s.db.Select("id, image_path, name").
		Where("id = ?", userId).First(&profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.Profile{}, ErrProfileNotFound
		}

		return &entity.Profile{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return profile, nil
}

func (s *storage) GetMe(userId string) (*entity.Profile, error) {
	const fn = "adapters.repository.GetMe"
	var profile *entity.Profile

	result := s.db.Where("id = ?", userId).First(&profile)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.Profile{}, ErrProfileNotFound
		}

		return &entity.Profile{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return profile, nil
}

func (s *storage) DeleteProfileByID(id string) error {
	const fn = "adapters.repository.DeleteProfileByID"

	result := s.db.Where("id = ?", id).Delete(&entity.Profile{})
	if result.Error != nil {
		return fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrProfileNotFound
	}

	return nil
}

func (s *storage) UpdateProfileFields(id string, updates map[string]interface{}) error {
	const fn = "adapters.repository.UpdateProfileFields"

	result := s.db.Model(&entity.Profile{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrProfileNotFound
	}

	return nil
}
