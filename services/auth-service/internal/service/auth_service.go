package service

import (
	"context"
	"errors"
	"fmt"

	"paylater/services/auth-service/internal/repository"
	"paylater/shared/auth"
)

// AuthService contains authentication business logic.
type AuthService struct {
	users         repository.UserRepository
	merchants     repository.MerchantRepository
	jwtSecret     string
	adminEmail    string
	adminPassword string
}

// NewAuthService wires credential repositories and JWT/admin settings.
func NewAuthService(
	users repository.UserRepository,
	merchants repository.MerchantRepository,
	jwtSecret string,
	adminEmail string,
	adminPassword string,
) *AuthService {
	return &AuthService{
		users:         users,
		merchants:     merchants,
		jwtSecret:     jwtSecret,
		adminEmail:    adminEmail,
		adminPassword: adminPassword,
	}
}

// RegisterRequest is the service-layer user registration payload.
type RegisterRequest struct {
	Name     string
	Email    string
	Password string
}

// Register creates a new user with a hashed password.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) error {
	exists, err := s.users.EmailExists(ctx, req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already exists")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	_, err = s.users.CreateUser(ctx, req.Name, req.Email, hashedPassword)
	return err
}

// LoginRequest is the service-layer user login payload.
type LoginRequest struct {
	Email    string
	Password string
}

// Login authenticates a user and returns a JWT with role "user".
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (string, error) {
	user, err := s.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	if err := auth.CheckPassword(user.Password, req.Password); err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(user.UserID, user.Email, "user", s.jwtSecret)
	if err != nil {
		return "", err
	}
	return token, nil
}

// AdminLoginRequest is the service-layer admin login payload.
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin authenticates admin credentials from environment config
// and returns a JWT with role "admin" and user_id 0.
func (s *AuthService) AdminLogin(ctx context.Context, req AdminLoginRequest) (string, error) {
	_ = ctx

	if req.Email != s.adminEmail || req.Password != s.adminPassword {
		return "", errors.New("invalid admin credentials")
	}

	token, err := auth.GenerateToken(0, s.adminEmail, "admin", s.jwtSecret)
	if err != nil {
		return "", err
	}
	return token, nil
}

// MerchantRegisterRequest is the merchant self-registration payload.
type MerchantRegisterRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Email                string  `json:"email" binding:"required,email"`
	Phone                string  `json:"phone" binding:"required"`
	Password             string  `json:"password" binding:"required,min=6"`
	CommissionPercentage float64 `json:"commission_percentage" binding:"required"`
}

// MerchantRegister creates a merchant with a hashed password.
func (s *AuthService) MerchantRegister(ctx context.Context, req MerchantRegisterRequest) error {
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	_, err = s.merchants.CreateMerchant(
		ctx,
		req.Name,
		req.Email,
		req.Phone,
		hashedPassword,
		fmt.Sprintf("%.2f", req.CommissionPercentage),
	)
	return err
}

// MerchantLoginRequest is the merchant login payload.
type MerchantLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// MerchantLogin authenticates a merchant and returns a JWT with role "merchant".
func (s *AuthService) MerchantLogin(ctx context.Context, req MerchantLoginRequest) (string, error) {
	merchant, err := s.merchants.GetMerchantByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	if err := auth.CheckPassword(merchant.PasswordHash, req.Password); err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := auth.GenerateToken(
		merchant.MerchantID,
		merchant.Email,
		"merchant",
		s.jwtSecret,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
