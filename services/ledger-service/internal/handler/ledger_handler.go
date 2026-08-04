package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"paylater/services/ledger-service/internal/repository"
	"paylater/services/ledger-service/internal/service"
	"paylater/shared/response"
)

// LedgerHandler exposes purchase and payment HTTP endpoints.
type LedgerHandler struct {
	service *service.LedgerService
}

// NewLedgerHandler creates a LedgerHandler.
func NewLedgerHandler(s *service.LedgerService) *LedgerHandler {
	return &LedgerHandler{service: s}
}

type purchaseRequest struct {
	MerchantID int32   `json:"merchant_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
}

type repayRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

type transactionResponse struct {
	TransactionID        int32     `json:"transaction_id"`
	UserID               int32     `json:"user_id"`
	MerchantID           int32     `json:"merchant_id"`
	Amount               string    `json:"amount"`
	CommissionPercentage string    `json:"commission_percentage"`
	CommissionAmount     string    `json:"commission_amount"`
	CreatedAt            time.Time `json:"created_at"`
}

type paymentResponse struct {
	PaymentID int32     `json:"payment_id"`
	UserID    int32     `json:"user_id"`
	Amount    string    `json:"amount"`
	PaidAt    time.Time `json:"paid_at"`
}

func toTx(t repository.Transaction) transactionResponse {
	return transactionResponse{
		TransactionID:        t.TransactionID,
		UserID:               t.UserID,
		MerchantID:           t.MerchantID,
		Amount:               t.Amount,
		CommissionPercentage: t.CommissionPercentage,
		CommissionAmount:     t.CommissionAmount,
		CreatedAt:            t.CreatedAt,
	}
}

func toPay(p repository.Payment) paymentResponse {
	return paymentResponse{
		PaymentID: p.PaymentID,
		UserID:    p.UserID,
		Amount:    p.Amount,
		PaidAt:    p.PaidAt,
	}
}

func jwtID(c *gin.Context) (int32, bool) {
	raw, ok := c.Get("id")
	if !ok {
		return 0, false
	}
	id, ok := raw.(int32)
	return id, ok
}

// Purchase handles POST /purchases.
func (h *LedgerHandler) Purchase(c *gin.Context) {
	userID, ok := jwtID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req purchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.service.Purchase(c.Request.Context(), userID, req.MerchantID, req.Amount)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Message(c, http.StatusCreated, "Purchase successful")
}

// Repay handles POST /payments.
func (h *LedgerHandler) Repay(c *gin.Context) {
	userID, ok := jwtID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req repayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.service.Repay(c.Request.Context(), userID, req.Amount)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Message(c, http.StatusOK, "Payment successful")
}

// ListMyPayments handles GET /payments for the authenticated user.
func (h *LedgerHandler) ListMyPayments(c *gin.Context) {
	userID, ok := jwtID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	payments, err := h.service.ListUserPayments(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]paymentResponse, 0, len(payments))
	for _, p := range payments {
		out = append(out, toPay(p))
	}
	response.JSON(c, http.StatusOK, out)
}

// ListPurchases handles GET /admin/purchases.
func (h *LedgerHandler) ListPurchases(c *gin.Context) {
	txs, err := h.service.ListTransactions(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]transactionResponse, 0, len(txs))
	for _, t := range txs {
		out = append(out, toTx(t))
	}
	response.JSON(c, http.StatusOK, out)
}

// GetPurchaseByID handles GET /admin/purchases/:id.
func (h *LedgerHandler) GetPurchaseByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid transaction id")
		return
	}
	tx, err := h.service.GetTransactionByID(c.Request.Context(), int32(id))
	if err != nil {
		writeErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toTx(tx))
}

// ListUserPurchases handles GET /admin/users/:id/purchases.
func (h *LedgerHandler) ListUserPurchases(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	txs, err := h.service.ListUserTransactions(c.Request.Context(), int32(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]transactionResponse, 0, len(txs))
	for _, t := range txs {
		out = append(out, toTx(t))
	}
	response.JSON(c, http.StatusOK, out)
}

// GetPaymentByID handles GET /admin/payments/:id.
func (h *LedgerHandler) GetPaymentByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payment id")
		return
	}
	p, err := h.service.GetPaymentByID(c.Request.Context(), int32(id))
	if err != nil {
		writeErr(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toPay(p))
}

// ListUserPaymentsAdmin handles GET /admin/users/:id/payments.
func (h *LedgerHandler) ListUserPaymentsAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	payments, err := h.service.ListUserPayments(c.Request.Context(), int32(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]paymentResponse, 0, len(payments))
	for _, p := range payments {
		out = append(out, toPay(p))
	}
	response.JSON(c, http.StatusOK, out)
}

// ListMerchantTransactions handles GET /merchant/transactions.
func (h *LedgerHandler) ListMerchantTransactions(c *gin.Context) {
	merchantID, ok := jwtID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	txs, err := h.service.ListMerchantTransactions(c.Request.Context(), merchantID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]transactionResponse, 0, len(txs))
	for _, t := range txs {
		out = append(out, toTx(t))
	}
	response.JSON(c, http.StatusOK, out)
}

// MerchantCommissions handles GET /internal/reports/merchant-commissions.
func (h *LedgerHandler) MerchantCommissions(c *gin.Context) {
	rows, err := h.service.MerchantCommissionSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"merchant_id":      r.MerchantID,
			"total_commission": r.TotalCommission,
		})
	}
	response.JSON(c, http.StatusOK, out)
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAmountMustBePositive),
		errors.Is(err, service.ErrInsufficientCredit),
		errors.Is(err, service.ErrPaymentExceedsDue):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrMerchantNotFound),
		errors.Is(err, service.ErrUserNotFound),
		errors.Is(err, service.ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrServiceUnavailable):
		response.Error(c, http.StatusServiceUnavailable, "service unavailable")
	default:
		response.Error(c, http.StatusInternalServerError, "internal error")
	}
}
