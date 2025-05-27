package api

import (
	"net/http"
	"sales-api/internal/sale"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user CRUD endpoints on the given Gin engine.
// It initializes the storage, service, and handler, then binds each HTTP
// method and path to the appropriate handler function.
func InitRoutes(e *gin.Engine, userAPIURL string) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	storage := sale.NewLocalStorage()
	saleService := sale.NewService(storage, logger, userAPIURL)

	h := handler{
		saleService: saleService,
		logger:      logger,
	}

	e.POST("/sales", h.handleCreateSale)
	e.GET("/sales", h.handleReadSale)
	e.PATCH("/sales/:id", h.handleUpdateSale)

	e.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}
