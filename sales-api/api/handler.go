package api

import (
	"errors"
	"net/http"
	"sales-api/internal/sale"
	"users-api/internal/user"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// handler holds the user service and implements HTTP handlers for user CRUD.
type handler struct {
	saleService *sale.Service
	userService *user.Service
	logger      *zap.Logger
}

func (h *handler) handleCreateSale(ctx *gin.Context) {
	// request payload
	var req struct {
		UserID string  `json:"user_id"`
		Amount float32 `json:"amount"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.userService.GetUser(req.UserID); err == sale.ErrNotFound {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &sale.Sale{
		UserID: req.UserID,
		Amount: req.Amount,
	}
	if err := h.saleService.CreateSale(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("sale created", zap.Any("sale", u))
	ctx.JSON(http.StatusCreated, u)
}

func (h *handler) handleReadSale(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	status := ctx.Query("status")

	u, err := h.saleService.GetSaleByUserAndStatus(userID, status)
	if err != nil {
		if errors.Is(err, sale.ErrInvalidInput) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get sale", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("get user succeed", zap.Any("user", u))
	ctx.JSON(http.StatusOK, u)
}

// handleUpdate handles PUT /users/:id
func (h *handler) handleUpdateSale(ctx *gin.Context) {
	id := ctx.Param("id")
	// bind partial update fields
	var fields *sale.UpdateFieldsSale
	if err := ctx.ShouldBindJSON(&fields); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.saleService.UpdateSale(id, fields)
	if err != nil {
		if errors.Is(err, sale.ErrNotFoundSale) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, sale.ErrInvalidInput) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, sale.ErrTransactionInvalid) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, u)
}
