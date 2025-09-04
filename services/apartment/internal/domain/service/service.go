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

	"airbnb.com/services/apartment/internal/adapters/repository"
	"airbnb.com/services/apartment/internal/domain/entity"
	"github.com/google/uuid"
)

type ApartmentService interface {
	CreateApartment(req *entity.CreateApartmentRequest, hostID string, imageFiles []*multipart.FileHeader) (*entity.ApartmentResponse, error)
	GetApartmentByID(id string) (*entity.ApartmentResponse, error)
	DeleteApartment(id string) error
	UpdateApartment(id string, updates map[string]interface{}, imageFiles []*multipart.FileHeader) (*entity.ApartmentResponse, error)
}

type apartmentService struct {
	repo      repository.ApartmentRepository
	uploadDir string
	log       *slog.Logger
}

func NewApartmentService(repo repository.ApartmentRepository, uploadDir string, log *slog.Logger) ApartmentService {
	return &apartmentService{repo: repo, uploadDir: uploadDir, log: log}
}

func (s *apartmentService) CreateApartment(req *entity.CreateApartmentRequest, hostID string, imageFiles []*multipart.FileHeader) (*entity.ApartmentResponse, error) {
	const fn = "domain.service.CreateApartment"
	log := s.log.With(slog.String("fn", fn))

	if req.Title == "" || req.PricePerNight <= 0 || hostID == "" {
		log.Error("failed create an apt. no required fields")
		return nil, fmt.Errorf("%s: %w", fn, ErrInvalidInput)
	}

	apt := &entity.Apartment{
		ID:            uuid.New().String(),
		HostID:        hostID,
		Title:         req.Title,
		Description:   req.Description,
		PricePerNight: req.PricePerNight,
		HouseNumber:   req.HouseNumber,
		Street:        req.Street,
		City:          req.City,
		State:         req.State,
		Country:       req.Country,
		PostalCode:    req.PostalCode,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		Wifi:          req.Wifi,
		Parking:       req.Parking,
		AirCondition:  req.AirCondition,
		Kitchen:       req.Kitchen,
		PetFriendly:   req.PetFriendly,
		MaxGuests:     req.MaxGuests,
		BedroomNumber: req.BedroomNumber,
	}

	var images []entity.Image
	for i, file := range imageFiles {
		path, err := s.saveImage(file, apt.ID)
		if err != nil {
			log.Error("failed to save image", slog.String("error", err.Error()))
			for _, img := range images {
				os.Remove(img.Path)
			}
			return nil, err
		}

		images = append(images, entity.Image{
			ID:          uuid.New().String(),
			ApartmentID: apt.ID,
			Path:        path,
			IsCover:     i == 0, // первая картинка = cover
		})
	}

	apt.Images = images

	if err := s.repo.CreateNewApartment(apt); err != nil {
		for _, img := range images {
			os.Remove(img.Path)
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	resp := toApartmentResponse(apt)
	return resp, nil
}

func (s *apartmentService) GetApartmentByID(id string) (*entity.ApartmentResponse, error) {
	const fn = "domain.service.GetApartmentByID"
	log := s.log.With(slog.String("fn", fn))

	apt, err := s.repo.GetApartment(id)
	if err != nil {
		if errors.Is(err, repository.ErrAptNotFound) {
			return nil, ErrAptNotFound
		}
		log.Error("failed to get apartment", slog.String("error", err.Error()))
		return nil, err
	}

	return toApartmentResponse(apt), nil
}

func (s *apartmentService) DeleteApartment(id string) error {
	const fn = "domain.service.DeleteApartment"
	log := s.log.With(slog.String("fn", fn))

	apt, err := s.repo.GetApartment(id)
	if err != nil {
		log.Error("failed to get an apt by its id", slog.String("error", err.Error()))
		return err
	}

	if err := s.repo.DeleteApartmentByID(id); err != nil {
		return err
	}

	for _, img := range apt.Images {
		os.Remove(img.Path)
	}

	return nil
}

func (s *apartmentService) UpdateApartment(id string, updates map[string]interface{}, imageFiles []*multipart.FileHeader) (*entity.ApartmentResponse, error) {
	const fn = "domain.service.UpdateApartment"
	log := s.log.With(slog.String("fn", fn))

	var newImages []entity.Image
	for i, file := range imageFiles {
		path, err := s.saveImage(file, id)
		if err != nil {
			log.Error("failed to save image", slog.String("error", err.Error()))
			for _, img := range newImages {
				os.Remove(img.Path)
			}
			return nil, err
		}
		newImages = append(newImages, entity.Image{
			ID:          uuid.New().String(),
			ApartmentID: id,
			Path:        path,
			IsCover:     i == 0,
		})
	}

	if len(updates) > 0 {
		if err := s.repo.UpdateApartmentFields(id, updates); err != nil {
			log.Error("failed to update apartment fields", slog.String("error", err.Error()))
			return nil, err
		}
	}

	if len(newImages) > 0 {
		if err := s.repo.AddImages(id, newImages); err != nil {
			for _, img := range newImages {
				os.Remove(img.Path)
			}
			return nil, err
		}
	}

	apt, err := s.repo.GetApartment(id)
	if err != nil {
		return nil, err
	}

	return toApartmentResponse(apt), nil
}

func (s *apartmentService) saveImage(file *multipart.FileHeader, apartmentID string) (string, error) {
	if file.Size > 5*1024*1024 {
		return "", ErrImageTooLarge
	}

	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		return "", ErrInvalidImage
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", apartmentID, time.Now().Unix(), ext)
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

func toApartmentResponse(apt *entity.Apartment) *entity.ApartmentResponse {
	resp := &entity.ApartmentResponse{
		ID:            apt.ID,
		HostID:        apt.HostID,
		Title:         apt.Title,
		Description:   apt.Description,
		PricePerNight: apt.PricePerNight,
		HouseNumber:   apt.HouseNumber,
		Street:        apt.Street,
		City:          apt.City,
		State:         apt.State,
		Country:       apt.Country,
		PostalCode:    apt.PostalCode,
		Latitude:      apt.Latitude,
		Longitude:     apt.Longitude,
		Wifi:          apt.Wifi,
		Parking:       apt.Parking,
		AirCondition:  apt.AirCondition,
		Kitchen:       apt.Kitchen,
		PetFriendly:   apt.PetFriendly,
		MaxGuests:     apt.MaxGuests,
		BedroomNumber: apt.BedroomNumber,
		CreatedAt:     apt.CreatedAt,
		UpdatedAt:     apt.UpdatedAt,
	}

	for _, img := range apt.Images {
		resp.Images = append(resp.Images, entity.ImageResponse{
			ID:      img.ID,
			URL:     "/uploads/" + filepath.Base(img.Path),
			IsCover: img.IsCover,
		})
	}

	return resp
}
