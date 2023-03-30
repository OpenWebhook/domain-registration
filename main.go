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
	Login string `json:"login" binding:"required"`
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

	dnsRecordContent := os.Getenv("DNS_RECORD_CONTENT")
	if dnsRecordContent == "" {
		log.Fatal("$DNS_RECORD_CONTENT must be set")
	}

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	cloudflareZoneID, err := api.ZoneIDByName(os.Getenv("CLOUDFLARE_DOMAIN"))
	if err != nil {
		log.Fatal(err)
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

		name := "*." + json.Sender.Login + ".github"

		fmt.Println("Attempting to create dns record for " + name)

		var createDNSRecordParams = cloudflare.CreateDNSRecordParams{
			Content: dnsRecordContent,
			Name:    name,
			Type:    "CNAME",
			Comment: "Github user wildcard",
			TTL:     3600,
		}

		record, err := api.CreateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(cloudflareZoneID), createDNSRecordParams)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusCreated, gin.H{"message": "star gazer not created"})
			return
		}
		fmt.Println(record)

		c.JSON(http.StatusCreated, gin.H{"message": "star gazer created"})

	})

	router.Run(":" + port)
}
