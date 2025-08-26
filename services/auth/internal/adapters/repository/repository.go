package repository

import (
	"errors"
	"fmt"

	"airbnb.com/services/auth/internal/config"
	domain "airbnb.com/services/auth/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AuthRepository interface {
	CreateNewUser(user *domain.UserCredentials) (string, error)
	GetUserByEmail(email string) (*domain.UserCredentials, error)
	CreateRefreshToken(token *domain.RefreshToken) error
	ValidateRefreshToken(tokenValue string) (domain.RefreshToken, error)
}

type storage struct {
	db *gorm.DB
}

func New(cfg *config.Config) (AuthRepository, error) {
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

	err = db.AutoMigrate(&domain.UserCredentials{}, &domain.RefreshToken{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}
	return &storage{db: db}, nil
}

func (s *storage) CreateNewUser(user *domain.UserCredentials) (string, error) {
	const fn = "adapters.repository.CreateNewUser"

	result := s.db.Create(&user)
	var pgErr *pgconn.PgError

	if result.Error != nil {
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return "", ErrEmailExist
		}
		return "", fmt.Errorf("%s: %w", fn, result.Error)
	}

	return user.ID, nil
}

func (s *storage) GetUserByEmail(email string) (*domain.UserCredentials, error) {
	const fn = "adapters.repository.GetUserByEmail"
	var user *domain.UserCredentials

	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &domain.UserCredentials{}, ErrEmailNotFound
		}

		return &domain.UserCredentials{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return user, nil
}

func (s *storage) CreateRefreshToken(token *domain.RefreshToken) error {
	const fn = "adapters.repository.CreateRefreshToken"

	result := s.db.Create(&token)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", fn, result.Error)
	}
	return nil
}

func (s *storage) ValidateRefreshToken(tokenValue string) (domain.RefreshToken, error) {
	const fn = "adapters.repository.ValidateRefreshToken"

	var token domain.RefreshToken

	result := s.db.Where("token_hash = ?", tokenValue).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return domain.RefreshToken{}, ErrRefreshTokenNotFound
		}

		return domain.RefreshToken{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return token, nil
}
