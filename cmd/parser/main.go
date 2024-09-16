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
	"github.com/manzanit0/mcduck/internal/parser"
	"github.com/manzanit0/mcduck/pkg/micro"
	"github.com/manzanit0/mcduck/pkg/openai"
	"github.com/manzanit0/mcduck/pkg/xtrace"
	"go.opentelemetry.io/otel/attribute"
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

	// Just crash the service if these aren't available.
	_ = micro.MustGetEnv("AWS_ACCESS_KEY")
	_ = micro.MustGetEnv("AWS_SECRET_ACCESS_KEY")

	config, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(awsRegion))
	if err != nil {
		panic(err)
	}

	textractParser := parser.NewTextractParser(config, apiKey)
	aivisionParser := parser.NewAIVisionParser(apiKey)

	svc.Engine.POST("/receipt", func(c *gin.Context) {
		ctx, span := xtrace.StartSpan(c.Request.Context(), "Parse Receipt")
		defer span.End()

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

		span.SetAttributes(attribute.String("file.path", filePath))

		data, err := os.ReadFile(filePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unable to read file from disk: %s", err.Error())})
			return
		}

		span.SetAttributes(attribute.Int("file.size", len(data)))

		contentType := http.DetectContentType(data)
		span.SetAttributes(attribute.String("file.content_type", contentType))

		var receipt *parser.Receipt
		var openAIRes *openai.Response

		switch contentType {
		case "application/pdf":
			receipt, openAIRes, err = textractParser.ExtractReceipt(ctx, data)
			if err != nil {
				marshalledRes, _ := json.Marshal(openAIRes)
				span.SetAttributes(attribute.String("openai.response", string(marshalledRes)))
				slog.ErrorContext(c.Request.Context(), "failed to extract receipt", "error", err.Error(), "open_ai_response", marshalledRes)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable extract data from receipt: %s", err.Error())})
				return
			}

		// Default to images
		default:
			receipt, openAIRes, err = aivisionParser.ExtractReceipt(ctx, data)
			if err != nil {
				marshalledRes, _ := json.Marshal(openAIRes)
				span.SetAttributes(attribute.String("openai.response", string(marshalledRes)))
				slog.ErrorContext(c.Request.Context(), "chatGPT response", "error", err.Error(), "open_ai_response", marshalledRes)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("unable extract data from receipt: %s", err.Error())})
				return
			}
		}

		marshalled, _ := json.Marshal(receipt)
		marshalledRes, _ := json.Marshal(openAIRes)
		span.SetAttributes(attribute.String("openai.response", string(marshalledRes)))
		slog.InfoContext(c.Request.Context(), "chatGPT response", "processed_receipt", marshalled, "open_ai_response", marshalledRes)

		c.JSON(http.StatusOK, receipt)
	})

	if err := svc.Run(); err != nil {
		slog.Error("run ended with error", "error", err.Error())
		os.Exit(1)
	}
}
