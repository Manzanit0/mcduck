package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/pkg/micro"
)

const (
	serviceName = "parser"
	awsRegion   = "eu-west-1"
)

func main() {
	svc, err := micro.NewGinService(serviceName)
	if err != nil {
		panic(err)
	}

	apiKey := micro.MustGetEnv("OPENAI_API_KEY")

	config, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	if err != nil {
		panic(err)
	}

	textractParser := NewTextractParser(config, apiKey)

	svc.Engine.POST("/receipt", func(c *gin.Context) {
		file, err := c.FormFile("receipt")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to read file: %s", err.Error())})
			return
		}

		dir := os.TempDir()
		filePath := filepath.Join(dir, file.Filename)
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to save file to disk: %s", err.Error())})
			return
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to read file from disk: %s", err.Error())})
			return
		}

		var response *Receipt
		switch http.DetectContentType(data) {
		case "application/pdf":
			response, err = textractParser.ExtractReceipt(c.Request.Context(), data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable extract data from receipt: %s", err.Error())})
				return
			}

		// Default to images
		default:
			response, err = parseReceiptImage(c.Request.Context(), apiKey, data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable extract data from receipt: %s", err.Error())})
				return
			}
		}

		marshalled, _ := json.Marshal(response)
		slog.InfoContext(c.Request.Context(), "chatGPT response", "body", marshalled)

		c.JSON(http.StatusOK, response)
	})

	if err := svc.Run(); err != nil {
		slog.Error("run ended with error", "error", err.Error())
		os.Exit(1)
	}
}
