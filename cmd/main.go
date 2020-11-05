package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"github.com/philLITERALLY/wodland-service/internal/http"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// Set up heroku database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Endpoint to get single WOD and any attempts at it
	router.GET("/WOD/:wodID", http.GetWOD(db))

	// Endpoint to get WODs (can be filtered)
	router.GET("/WODs", http.GetWODs(db))

	// Endpoint to get WODs (can be filtered)
	router.GET("/Activities", http.GetActivities(db))

	// Endpoint to create a WOD (and add an attempt if supplied)
	router.POST("/WOD", http.AddWOD(db))

	// Endpoint to add an Activity
	router.POST("/Activity", http.AddActivity(db))

	router.Run(":" + port)
}
