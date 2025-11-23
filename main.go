package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"BACKEND-UAS/config"
)

func main() {
	cfg := config.NewConfig()

	// Log koneksi doang (dari connection.go)
	log.Printf("Connected to Postgres & Mongo – Ready!")

	r := gin.New() // New, no default middleware untuk bersih

	// Set mode release (hilangkan warning)
	gin.SetMode(gin.ReleaseMode)

	// Health check sederhana
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK", "db": "Connected"})
	})

	port := cfg.Port
	log.Printf("Server on port %s", port)
	r.Run(":" + port)
}