package service

import (
	"airbnb-clone/profile/internal/adapters/repository"
	"airbnb-clone/profile/internal/domain/entity"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type ProfileService interface {
	CreateProfile(request *entity.CreateProfileRequest, userId string, imageFile *multipart.FileHeader) (*entity.ProfileResponse, error)
	GetYourProfile(userId string) (*entity.ProfileResponse, error)
	GetProfile(userId string) (*entity.PublicProfileResponse, error)
	DeleteProfile(userId string) error
	UpdateProfile(userId string, request *entity.UpdateProfileRequest, imageFile *multipart.FileHeader) (*entity.ProfileResponse, error)
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

func (s *profileService) GetYourProfile(userId string) (*entity.ProfileResponse, error) {
	const fn = "domain.service.GetYourProfile"
	log := s.log.With(
		slog.String("fn", fn),
	)

	profile, err := s.profileRepository.GetMe(userId)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return &entity.ProfileResponse{}, ErrProfileNotFound
		}
		log.Error("failed to get me profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &entity.ProfileResponse{}, err
	}

	return &entity.ProfileResponse{
		ID:          profile.ID,
		PhoneNumber: profile.PhoneNumber,
		Name:        profile.Name,
		Surname:     profile.Surname,
		DateOfBirth: profile.DateOfBirth,
		ImageURL:    "/uploads/" + filepath.Base(profile.ImagePath),
	}, nil

}

func (s *profileService) GetProfile(userId string) (*entity.PublicProfileResponse, error) {
	const fn = "domain.service.GetProfile"
	log := s.log.With(
		slog.String("fn", fn),
	)

	profile, err := s.profileRepository.GetUserById(userId)
	if err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			return &entity.PublicProfileResponse{}, ErrProfileNotFound
		}
		log.Error("failed to get profile by user id", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &entity.PublicProfileResponse{}, err
	}

	return &entity.PublicProfileResponse{
		ID:       profile.ID,
		Name:     profile.Name,
		ImageURL: "/uploads/" + filepath.Base(profile.ImagePath),
	}, nil
}

func (s *profileService) DeleteProfile(userId string) error {
	const fn = "domain.service.DeleteProfile"
	log := s.log.With(
		slog.String("fn", fn),
	)

	if err := s.profileRepository.DeleteProfileByID(userId); err != nil {
		if errors.Is(err, repository.ErrProfileNotFound) {
			log.Error("profile not found", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return ErrProfileNotFound
		}
		log.Error("failed to delete user profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return err
	}

	return nil
}

func (s *profileService) UpdateProfile(userId string, request *entity.UpdateProfileRequest, imageFile *multipart.FileHeader) (*entity.ProfileResponse, error) {
	const fn = "domain.service.UpdateProfile"
	log := s.log.With(
		slog.String("fn", fn),
	)

	updates := make(map[string]interface{})

	if request.PhoneNumber != nil {
		updates["phone_number"] = *request.PhoneNumber
	}
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.Surname != nil {
		updates["surname"] = *request.Surname
	}
	if request.DateOfBirth != nil {
		updates["date_of_birth"] = *request.DateOfBirth
	}

	var newImagePath string
	if imageFile != nil {
		newImagePath, err := s.saveImage(imageFile, userId)
		if err != nil {
			log.Error("failed to save new image", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return nil, err
		}
		updates["image_path"] = newImagePath
	}

	if len(updates) > 0 {
		if err := s.profileRepository.UpdateProfileFields(userId, updates); err != nil {
			if newImagePath != "" {
				os.Remove(newImagePath)
			}
			log.Error("failed to update profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return nil, err
		}
	}

	updatedProfile, err := s.profileRepository.GetMe(userId)
	if err != nil {
		log.Error("failed to get updated profile", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return nil, err
	}

	return &entity.ProfileResponse{
		ID:          updatedProfile.ID,
		PhoneNumber: updatedProfile.PhoneNumber,
		Name:        updatedProfile.Name,
		Surname:     updatedProfile.Surname,
		DateOfBirth: updatedProfile.DateOfBirth,
		ImageURL:    "/uploads/" + filepath.Base(updatedProfile.ImagePath),
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
