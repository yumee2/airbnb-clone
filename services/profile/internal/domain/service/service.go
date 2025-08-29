package service

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"airbnb.com/services/profile/internal/adapters/repository"
	"airbnb.com/services/profile/internal/domain/entity"
)

type ProfileService interface {
	CreateProfile(request *entity.CreateProfileRequest, userId string, imageFile *multipart.FileHeader) (*entity.ProfileResponse, error)
}

type profileService struct {
	profileRepository repository.ProfileRepository
	log               *slog.Logger
	uploadDir         string
}

func NewProfileService(profileRepo repository.ProfileRepository, logger *slog.Logger, uploadDir string) ProfileService {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Error("Failed to create upload directory: %v", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		os.Exit(1)
	}

	return &profileService{profileRepository: profileRepo, log: logger, uploadDir: uploadDir}
}

func (s *profileService) CreateProfile(request *entity.CreateProfileRequest, userId string, imageFile *multipart.FileHeader) (*entity.ProfileResponse, error) {
	const fn = "domain.service.CreateProfile"
	log := s.log.With(
		slog.String("fn", fn),
	)

	imagePath, err := s.saveImage(imageFile, userId)
	if err != nil {
		log.Error("failed to save the image", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return nil, fmt.Errorf("error saving image: %w", err)
	}

	profile := &entity.Profile{
		ID:          userId,
		PhoneNumber: request.PhoneNumber,
		Name:        request.Name,
		Surname:     request.Surname,
		DateOfBirth: request.DateOfBirth,
		ImagePath:   imagePath,
	}

	if err := s.profileRepository.CreateNewProfile(profile); err != nil {
		os.Remove(imagePath)
		log.Error("failed to create a new profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return nil, fmt.Errorf("error creating profile: %w", err)
	}

	return &entity.ProfileResponse{
		ID:          profile.ID,
		PhoneNumber: profile.PhoneNumber,
		Name:        profile.Name,
		Surname:     profile.Surname,
		DateOfBirth: profile.DateOfBirth,
		ImageURL:    "/uploads/" + filepath.Base(imagePath),
	}, nil
}

func (s *profileService) saveImage(file *multipart.FileHeader, userID string) (string, error) {
	if file.Size > 5*1024*1024 { // 5MB limit
		return "", errors.New("image size too large")
	}

	allowedTypes := map[string]bool{ // validate MIME type
		"image/jpeg": true,
		"image/png":  true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		return "", errors.New("invalid image format")
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(s.uploadDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return filePath, nil
}
