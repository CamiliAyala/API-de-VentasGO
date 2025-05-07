package main

import (
	"fmt"
	"users-api/api"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	api.InitRoutes(r)

	if err := r.Run(":8080"); err != nil {
		panic(fmt.Errorf("error trying to start server: %v", err))
	}
}
