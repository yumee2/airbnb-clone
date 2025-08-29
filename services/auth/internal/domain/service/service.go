package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"time"

	"airbnb.com/services/auth/internal/domain/entity"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"airbnb.com/services/auth/internal/adapters/repository"
)

type AuthService interface {
	RegisterNewUser(email string, password string) (*JWTTokenPair, error)
	LoginExistingUser(email string, password string) (*JWTTokenPair, error)
	ValidateRefreshToken(refreshToken string) (string, error)
}

type authService struct {
	authRepository repository.AuthRepository
	log            *slog.Logger
}

type JWTTokenPair struct {
	RefreshToken       string
	AccessToken        string
	AccessExpireTime   time.Time
	RefreshExprireTime time.Time
}

func NewAuthService(authRepo repository.AuthRepository, logger *slog.Logger) AuthService {
	return &authService{authRepository: authRepo, log: logger}
}

// Return generated access and refresh tokens or error
func (s *authService) RegisterNewUser(email string, password string) (*JWTTokenPair, error) {
	const fn = "domain.service.RegisterNewUser"
	log := s.log.With(
		slog.String("fn", fn),
	)

	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Error("failed to hash password", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	newUser := &entity.UserCredentials{Email: email, Password: hashedPassword}

	uuid, err := s.authRepository.CreateNewUser(newUser)
	if err != nil {
		if errors.Is(err, repository.ErrEmailExist) {
			return &JWTTokenPair{}, ErrEmailExist
		}
		log.Error("failed to save new user credentials", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	jwtTokens, err := generateJWTTokenPair(uuid)
	if err != nil {
		log.Error("failed to create JWT tokens", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())},
			slog.Attr{Key: "UserUUID", Value: slog.StringValue(uuid)})
		return &JWTTokenPair{}, err
	}

	refreshToken := &entity.RefreshToken{TokenHash: jwtTokens.RefreshToken, UserID: uuid, ExpiresAt: jwtTokens.RefreshExprireTime}
	refreshToken.HashToken(jwtTokens.RefreshToken) // hashing the token to store in database
	if err = s.authRepository.CreateRefreshToken(refreshToken); err != nil {
		log.Error("failed to store refresh token in database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	return jwtTokens, nil
}

// Return  generated access and refresh tokens or error. Error can be ErrEmailNotFound type
func (s *authService) LoginExistingUser(email string, password string) (*JWTTokenPair, error) {
	const fn = "domain.service.LoginExistingUser"
	log := s.log.With(
		slog.String("fn", fn),
	)

	user, err := s.authRepository.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrEmailNotFound) {
			log.Error("user with provided email not found", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return &JWTTokenPair{}, ErrEmailNotFound
		}
		log.Error("failed to get user by email", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Error("invalid password", slog.Attr{Key: "email", Value: slog.StringValue(email)})
			return &JWTTokenPair{}, ErrInvalidPassword
		}
		log.Error("failed to compare password", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	jwtTokens, err := generateJWTTokenPair(user.ID)
	if err != nil {
		log.Error("failed to create JWT tokens", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())},
			slog.Attr{Key: "UserUUID", Value: slog.StringValue(user.ID)})
		return &JWTTokenPair{}, err
	}

	refreshToken := &entity.RefreshToken{TokenHash: jwtTokens.RefreshToken, UserID: user.ID, ExpiresAt: jwtTokens.RefreshExprireTime}
	refreshToken.HashToken(jwtTokens.RefreshToken) // hashing the token to store in database
	if err = s.authRepository.CreateRefreshToken(refreshToken); err != nil {
		log.Error("failed to store refresh token in database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	return jwtTokens, nil
}

// Return newly generated access token or error
func (s *authService) ValidateRefreshToken(refreshToken string) (string, error) {
	const fn = "domain.service.ValidateRefreshToken"
	log := s.log.With(
		slog.String("fn", fn),
	)

	hashedToken := hashToken(refreshToken)
	refresh, err := s.authRepository.ValidateRefreshToken(hashedToken)
	if err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotFound) {
			log.Error("provided token not found", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return "", ErrRefreshTokenNotFound
		}
		log.Error("failed to get a refresh token from database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return "", err
	}

	if !refresh.IsValid() { // check if the token is expired or not
		return "", ErrRefreshTokenExpired
	}

	jwtTokens, err := generateJWTTokenPair(refresh.UserID)
	if err != nil {
		log.Error("failed to create JWT tokens", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())},
			slog.Attr{Key: "UserUUID", Value: slog.StringValue(refresh.UserID)})
		return "", err
	}

	return jwtTokens.AccessToken, nil
}

func generateJWTTokenPair(userUUID string) (*JWTTokenPair, error) {
	var (
		jwtSecret  = os.Getenv("JWT_SECRET")
		accessTTL  = 15 * time.Minute
		refreshTTL = 7 * 24 * time.Hour
	)

	accessExpire := time.Now().Add(accessTTL)
	refreshExpire := time.Now().Add(refreshTTL)

	accessClaims := jwt.MapClaims{
		"user_id": userUUID,
		"exp":     accessExpire.Unix(),
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userUUID,
		"exp":     refreshExpire.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return &JWTTokenPair{}, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return &JWTTokenPair{}, err
	}

	return &JWTTokenPair{AccessToken: accessTokenString, RefreshToken: refreshTokenString,
		RefreshExprireTime: refreshExpire, AccessExpireTime: accessExpire}, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // bcrypt.DefaultCost is a good starting point
	return string(bytes), err
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])
	return hashedToken
}
