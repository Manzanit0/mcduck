package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/pkg/micro"
)

const (
	serviceName = "parser"
)

func main() {
	svc, err := micro.NewGinService(serviceName)
	if err != nil {
		panic(err)
	}

	apiKey := micro.MustGetEnv("OPENAI_API_KEY")

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
			c.JSON(http.StatusBadRequest, gin.H{"error": "PDF receipts are not supported"})
			return

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
