package api

import (
	"net/http"
	"parte3/internal/user"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user CRUD endpoints on the given Gin engine.
// It initializes the storage, service, and handler, then binds each HTTP
// method and path to the appropriate handler function.
func InitRoutes(e *gin.Engine) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	storage := user.NewLocalStorage()
	service := user.NewService(storage, logger)

	h := handler{
		userService: service,
		logger:      logger,
	}

	e.POST("/users", h.handleCreate)
	e.POST("/sales", h.handleCreateSale)
	e.GET("/users/:id", h.handleRead)
	e.GET("/sales", h.handleReadSale)
	e.PATCH("/users/:id", h.handleUpdate)
	e.PATCH("/sales/:id", h.handleUpdateSale)
	e.DELETE("/users/:id", h.handleDelete)

	e.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}
