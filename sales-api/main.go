package main

import (
	"fmt"
	"sales-api/api"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	api.InitRoutes(r, "http://localhost:8080")

	if err := r.Run(":8081"); err != nil {
		panic(fmt.Errorf("error trying to start server: %v", err))
	}
}
