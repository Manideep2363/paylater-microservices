package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"paylater/services/merchant-service/internal/repository"
	"paylater/services/merchant-service/internal/service"
	"paylater/shared/response"
)

// MerchantHandler exposes merchant HTTP endpoints.
type MerchantHandler struct {
	service *service.MerchantService
}

// NewMerchantHandler creates a MerchantHandler.
func NewMerchantHandler(s *service.MerchantService) *MerchantHandler {
	return &MerchantHandler{service: s}
}

type publicMerchantResponse struct {
	MerchantID           int32  `json:"merchant_id"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	CommissionPercentage string `json:"commission_percentage"`
}

type internalMerchantResponse struct {
	MerchantID           int32  `json:"merchant_id"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Phone                string `json:"phone"`
	PasswordHash         string `json:"password_hash"`
	CommissionPercentage string `json:"commission_percentage"`
}

func toPublic(m repository.MerchantView) publicMerchantResponse {
	return publicMerchantResponse{
		MerchantID:           m.MerchantID,
		Name:                 m.Name,
		Email:                m.Email,
		Phone:                m.Phone,
		CommissionPercentage: m.CommissionPercentage,
	}
}

func toInternal(m repository.MerchantInternal) internalMerchantResponse {
	return internalMerchantResponse{
		MerchantID:           m.MerchantID,
		Name:                 m.Name,
		Email:                m.Email,
		Phone:                m.Phone,
		PasswordHash:         m.PasswordHash,
		CommissionPercentage: m.CommissionPercentage,
	}
}

type createMerchantRequest struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required,email"`
	Phone      string  `json:"phone" binding:"required"`
	Password   string  `json:"password" binding:"required,min=6"`
	Commission float64 `json:"commission" binding:"required"`
}

type updateCommissionRequest struct {
	Commission float64 `json:"commission" binding:"required"`
}

// GetProfile handles GET /merchant/profile.
// Merchant ID comes only from the JWT; never from the client.
func (h *MerchantHandler) GetProfile(c *gin.Context) {
	rawID, exists := c.Get("id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "merchant id not found in token")
		return
	}
	merchantID, ok := rawID.(int32)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "invalid merchant id in token")
		return
	}

	merchant, err := h.service.GetMerchantByID(c.Request.Context(), merchantID)
	if err != nil {
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, toPublic(merchant))
}

// CreateMerchant handles POST /admin/merchants.
func (h *MerchantHandler) CreateMerchant(c *gin.Context) {
	var req createMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	merchant, err := h.service.CreateMerchant(
		c.Request.Context(),
		req.Name,
		req.Email,
		req.Phone,
		req.Password,
		req.Commission,
	)
	if err != nil {
		writeCreateError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toPublic(merchant))
}

// ListMerchants handles GET /admin/merchants.
func (h *MerchantHandler) ListMerchants(c *gin.Context) {
	merchants, err := h.service.ListMerchants(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]publicMerchantResponse, 0, len(merchants))
	for _, m := range merchants {
		out = append(out, toPublic(m))
	}
	response.JSON(c, http.StatusOK, out)
}

// GetMerchantByID handles GET /admin/merchants/:id.
func (h *MerchantHandler) GetMerchantByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid merchant id")
		return
	}

	merchant, err := h.service.GetMerchantByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Error(c, http.StatusNotFound, "merchant not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, toPublic(merchant))
}

// UpdateMerchantCommission handles PUT /admin/merchants/:id/commission.
func (h *MerchantHandler) UpdateMerchantCommission(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid merchant id")
		return
	}

	var req updateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.UpdateMerchantCommission(c.Request.Context(), int32(id), req.Commission)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCommission) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	response.Message(c, http.StatusOK, "Merchant commission updated successfully")
}

// InternalCreateMerchant handles POST /internal/merchants.
func (h *MerchantHandler) InternalCreateMerchant(c *gin.Context) {
	var req createMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	merchant, err := h.service.CreateMerchant(
		c.Request.Context(),
		req.Name,
		req.Email,
		req.Phone,
		req.Password,
		req.Commission,
	)
	if err != nil {
		writeCreateError(c, err)
		return
	}

	internal, err := h.service.GetMerchantByEmail(c.Request.Context(), merchant.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, toInternal(internal))
}

// InternalGetMerchantByID handles GET /internal/merchants/:id.
func (h *MerchantHandler) InternalGetMerchantByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid merchant id")
		return
	}

	merchant, err := h.service.GetMerchantByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	internal, err := h.service.GetMerchantByEmail(c.Request.Context(), merchant.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toInternal(internal))
}

// InternalGetMerchantByEmail handles GET /internal/merchants/by-email?email=...
func (h *MerchantHandler) InternalGetMerchantByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "email query parameter is required")
		return
	}

	merchant, err := h.service.GetMerchantByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, toInternal(merchant))
}

func writeCreateError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailExists),
		errors.Is(err, service.ErrInvalidCommission):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusBadRequest, err.Error())
	}
}
