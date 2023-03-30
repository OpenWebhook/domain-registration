package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudflare/cloudflare-go"
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

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	cloudflareZoneID, err := api.ZoneIDByName(os.Getenv("CLOUDFLARE_DOMAIN"))
	if err != nil {
		log.Fatal(err)
	}

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

		records, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(cloudflareZoneID), cloudflare.ListDNSRecordsParams{})
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, r := range records {
			fmt.Printf("%s: %s\n", r.Name, r.Content)
		}

		c.JSON(http.StatusCreated, gin.H{"message": "star gazer created", "numberOfRecords": len(records)})

	})

	router.Run(":" + port)
}
