package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

type Sender struct {
	Login string `json:"login"`
}

type JSONGithubWebhook struct {
	Sender Sender `json:"sender" binding:"required"`
	Action string `json:"action" binding:"required"`
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	router.POST("/", func(c *gin.Context) {
		var json JSONGithubWebhook
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.Action != "created" {
			c.JSON(http.StatusOK, gin.H{"message": "only created events are accepted"})
			return
		}
		log.Println(json.Sender.Login)
		c.JSON(http.StatusCreated, gin.H{"message": "star gazer created"})
	})

	router.Run(":" + port)
}
