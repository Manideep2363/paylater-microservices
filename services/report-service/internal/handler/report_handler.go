package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"paylater/services/report-service/internal/service"
	"paylater/shared/response"
)

// ReportHandler exposes admin report endpoints.
type ReportHandler struct {
	service *service.ReportService
}

// NewReportHandler creates a ReportHandler.
func NewReportHandler(s *service.ReportService) *ReportHandler {
	return &ReportHandler{service: s}
}

// OutstandingBalance handles GET /admin/reports/outstanding-balance.
func (h *ReportHandler) OutstandingBalance(c *gin.Context) {
	total, err := h.service.OutstandingBalance(c.Request.Context())
	if err != nil {
		writeErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, gin.H{
		"total_outstanding_balance": total,
	})
}

// UsersDue handles GET /admin/reports/users-due.
func (h *ReportHandler) UsersDue(c *gin.Context) {
	rows, err := h.service.UsersDue(c.Request.Context())
	if err != nil {
		writeErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, rows)
}

// UsersAtCreditLimit handles GET /admin/reports/users-at-credit-limit.
func (h *ReportHandler) UsersAtCreditLimit(c *gin.Context) {
	rows, err := h.service.UsersAtCreditLimit(c.Request.Context())
	if err != nil {
		writeErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, rows)
}

// MerchantCommissions handles GET /admin/reports/merchant-commissions.
func (h *ReportHandler) MerchantCommissions(c *gin.Context) {
	rows, err := h.service.MerchantCommissions(c.Request.Context())
	if err != nil {
		writeErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, rows)
}

func writeErr(c *gin.Context, err error) {
	if errors.Is(err, service.ErrServiceUnavailable) {
		response.Error(c, http.StatusServiceUnavailable, "service unavailable")
		return
	}
	response.Error(c, http.StatusInternalServerError, "internal error")
}
