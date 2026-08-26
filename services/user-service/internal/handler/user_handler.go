package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"paylater/services/user-service/internal/repository"
	"paylater/services/user-service/internal/service"
	"paylater/shared/response"
)

// UserHandler exposes user HTTP endpoints.
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a UserHandler.
func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

type publicUserResponse struct {
	UserID      int32  `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	CreditLimit string `json:"credit_limit"`
	CurrentDue  string `json:"current_due"`
}

type internalUserResponse struct {
	UserID      int32  `json:"user_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	CreditLimit string `json:"credit_limit"`
	CurrentDue  string `json:"current_due"`
}

func toPublic(u repository.UserView) publicUserResponse {
	return publicUserResponse{
		UserID:      u.UserID,
		Name:        u.Name,
		Email:       u.Email,
		CreditLimit: u.CreditLimit,
		CurrentDue:  u.CurrentDue,
	}
}

func toInternal(u repository.UserInternal) internalUserResponse {
	return internalUserResponse{
		UserID:      u.UserID,
		Name:        u.Name,
		Email:       u.Email,
		Password:    u.Password,
		CreditLimit: u.CreditLimit,
		CurrentDue:  u.CurrentDue,
	}
}

type createUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type dueRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

// GetUserByID handles GET /users/:id.
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)
	rawID, _ := c.Get("id")
	callerUserID, _ := rawID.(int32)

	if role != "admin" && callerUserID != int32(id) {
		response.Error(c, http.StatusForbidden, "access denied")
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, toPublic(user))
}

// ListUsers handles GET /admin/users.
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]publicUserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, toPublic(u))
	}
	response.JSON(c, http.StatusOK, out)
}

// CreateUser handles POST /admin/users.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusCreated, toPublic(user))
}

// InternalCreateUser handles POST /internal/users.
func (h *UserHandler) InternalCreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Internal create returns password hash for auth-service consumption later.
	internal, err := h.service.GetUserByEmail(c.Request.Context(), user.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, toInternal(internal))
}

// InternalGetUserByID handles GET /internal/users/:id.
func (h *UserHandler) InternalGetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), int32(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	internal, err := h.service.GetUserByEmail(c.Request.Context(), user.Email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toInternal(internal))
}

// InternalGetUserByEmail handles GET /internal/users/by-email?email=...
func (h *UserHandler) InternalGetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "email query parameter is required")
		return
	}

	user, err := h.service.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(c, http.StatusOK, toInternal(user))
}

// IncreaseDue handles POST /internal/users/:id/due/increase.
func (h *UserHandler) IncreaseDue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.IncreaseDue(c.Request.Context(), int32(id), req.Amount)
	if err != nil {
		writeDueError(c, err)
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), int32(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toPublic(user))
}

// DecreaseDue handles POST /internal/users/:id/due/decrease.
func (h *UserHandler) DecreaseDue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req dueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.DecreaseDue(c.Request.Context(), int32(id), req.Amount)
	if err != nil {
		writeDueError(c, err)
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), int32(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toPublic(user))
}

func writeDueError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrAmountMustBePositive),
		errors.Is(err, service.ErrInsufficientCredit),
		errors.Is(err, service.ErrPaymentExceedsDue):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}

// OutstandingBalance handles GET /internal/reports/outstanding-balance.
func (h *UserHandler) OutstandingBalance(c *gin.Context) {
	total, err := h.service.GetOutstandingBalance(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"total_outstanding_balance": total,
	})
}

// UserOutstandingDues handles GET /internal/reports/users-due.
func (h *UserHandler) UserOutstandingDues(c *gin.Context) {
	rows, err := h.service.GetUserOutstandingDues(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"user_id":     r.UserID,
			"name":        r.Name,
			"current_due": r.CurrentDue,
		})
	}
	response.JSON(c, http.StatusOK, out)
}

// UsersAtCreditLimit handles GET /internal/reports/users-at-credit-limit.
func (h *UserHandler) UsersAtCreditLimit(c *gin.Context) {
	rows, err := h.service.GetUsersAtCreditLimit(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"user_id":      r.UserID,
			"name":         r.Name,
			"credit_limit": r.CreditLimit,
			"current_due":  r.CurrentDue,
		})
	}
	response.JSON(c, http.StatusOK, out)
}
