package api

import (
	"errors"
	"net/http"
	"parte3/internal/user"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// handler holds the user service and implements HTTP handlers for user CRUD.
type handler struct {
	userService *user.Service
	logger      *zap.Logger
}

// handleCreate handles POST /users
func (h *handler) handleCreate(ctx *gin.Context) {
	// request payload
	var req struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		NickName string `json:"nickname"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &user.User{
		Name:     req.Name,
		Address:  req.Address,
		NickName: req.NickName,
	}
	if err := h.userService.CreateUser(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user created", zap.Any("user", u))
	ctx.JSON(http.StatusCreated, u)
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
	if _, err := h.userService.GetUser(req.UserID); err == user.ErrNotFound {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &user.Sale{
		UserID: req.UserID,
		Amount: req.Amount,
	}
	if err := h.userService.CreateSale(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user created", zap.Any("user", u))
	ctx.JSON(http.StatusCreated, u)
}

// handleRead handles GET /users/:id
// Si el usuario no existe: 400 bad request.
func (h *handler) handleRead(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := h.userService.GetUser(id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.logger.Warn("user not found", zap.String("id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("get user succeed", zap.Any("user", u))
	ctx.JSON(http.StatusOK, u)
}
func (h *handler) handleReadSale(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	status := ctx.Query("status")

	u, err := h.userService.GetSaleByUserAndStatus(userID, status)
	if err != nil {
		if errors.Is(err, user.ErrInvalidInput) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("get user succeed", zap.Any("user", u))
	ctx.JSON(http.StatusOK, u)
}

// handleUpdate handles PUT /users/:id
func (h *handler) handleUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	// bind partial update fields
	var fields *user.UpdateFieldsUser
	if err := ctx.ShouldBindJSON(&fields); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userService.UpdateUser(id, fields)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, u)
}

func (h *handler) handleUpdateSale(ctx *gin.Context) {
	id := ctx.Param("id")
	// bind partial update fields
	var fields *user.UpdateFieldsSale
	if err := ctx.ShouldBindJSON(&fields); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userService.UpdateSale(id, fields)
	if err != nil {
		if errors.Is(err, user.ErrNotFoundSale) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, user.ErrInvalidInput) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, user.ErrTransactionInvalid) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, u)
}

// handleDelete handles DELETE /users/:id
func (h *handler) handleDelete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := h.userService.Delete(id); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
