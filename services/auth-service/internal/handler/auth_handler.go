package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"paylater/services/auth-service/internal/repository"
	"paylater/services/auth-service/internal/service"
	"paylater/shared/response"
)

// AuthHandler exposes authentication HTTP endpoints.
type AuthHandler struct {
	service *service.AuthService
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// Register handles POST /register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.service.Register(c.Request.Context(), service.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err, http.StatusBadRequest)
		return
	}

	response.Message(c, http.StatusCreated, "User registered successfully")
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.service.Login(c.Request.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeAuthError(c, err, http.StatusUnauthorized)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"token": token})
}

// AdminLogin handles POST /admin/login.
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req service.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.service.AdminLogin(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"token": token})
}

// MerchantRegister handles POST /merchant/register.
func (h *AuthHandler) MerchantRegister(c *gin.Context) {
	var req service.MerchantRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.MerchantRegister(c.Request.Context(), req); err != nil {
		writeAuthError(c, err, http.StatusInternalServerError)
		return
	}

	response.Message(c, http.StatusCreated, "merchant registered successfully")
}

// MerchantLogin handles POST /merchant/login.
func (h *AuthHandler) MerchantLogin(c *gin.Context) {
	var req service.MerchantLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.service.MerchantLogin(c.Request.Context(), req)
	if err != nil {
		writeAuthError(c, err, http.StatusUnauthorized)
		return
	}

	response.JSON(c, http.StatusOK, gin.H{"token": token})
}

func writeAuthError(c *gin.Context, err error, fallbackStatus int) {
	if errors.Is(err, repository.ErrUnavailable) {
		response.Error(c, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	response.Error(c, fallbackStatus, err.Error())
}
