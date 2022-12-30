package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/pkg/invx"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/olekukonko/tablewriter"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	invxClient := invx.NewClient(os.Getenv("INVX_HOST"), os.Getenv("INVX_AUTH_TOKEN"))
	tgramClient := tgram.NewClient(http.DefaultClient, os.Getenv("TELEGRAM_BOT_TOKEN"))

	r.POST("/telegram/webhook", telegramWebhookController(tgramClient, invxClient))

	// background job to ping users on weather changes
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: r}
	go func() {
		log.Printf("serving HTTP on :%s", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server ended abruptly: %s", err.Error())
		} else {
			log.Printf("server ended gracefully")
		}

		stop()
	}()

	// Listen for OS interrupt
	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}

	log.Printf("server exited")
}

// @see https://core.telegram.org/bots/api#markdownv2-style
func webhookResponse(p *tgram.WebhookRequest, text string) gin.H {
	return gin.H{
		"method":     "sendMessage",
		"chat_id":    p.GetFromID(),
		"text":       text,
		"parse_mode": "MarkdownV2",
	}
}

func telegramWebhookController(tgramClient tgram.Client, invxClient invx.Client) func(c *gin.Context) {
	return func(c *gin.Context) {
		var r tgram.WebhookRequest
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Errorf("payload does not conform with telegram contract: %w", err).Error(),
			})
			return
		}

		if r.Message == nil || len(r.Message.Photos) == 0 {
			c.JSON(http.StatusOK, webhookResponse(&r, "Hey! Just send me a picture with a receipt ;-)"))
			return
		}

		// Get the biggest photo: this will ensure better parsing by invx service.
		var fileID string
		var fileSize int64
		for _, p := range r.Message.Photos {
			if p.FileSize != nil && *p.FileSize > fileSize {
				fileID = p.FileID
				fileSize = *p.FileSize
			}
		}

		file, err := tgramClient.GetFile(tgram.GetFileRequest{FileID: fileID})
		if err != nil {
			c.JSON(http.StatusOK, webhookResponse(&r, fmt.Sprintf("unable to get file from Telegram servers: %s", err.Error())))
			return
		}

		fileData, err := tgramClient.DownloadFile(file)
		if err != nil {
			c.JSON(http.StatusOK, webhookResponse(&r, fmt.Sprintf("unable to download file from Telegram servers: %s", err.Error())))
			return
		}

		if len(fileData) == 0 {
			c.JSON(http.StatusOK, webhookResponse(&r, "empty file"))
			return
		}

		amounts, err := invxClient.ParseReceipt(c.Request.Context(), fileData)
		if err != nil {
			c.JSON(http.StatusOK, webhookResponse(&r, fmt.Sprintf("unable to parser receipt: %s", err.Error())))
			return
		}

		c.JSON(http.StatusOK, webhookResponse(&r, NewBreakdownTgramMessage(amounts)))
	}
}

func NewBreakdownTgramMessage(amounts map[string]float64) string {
	b := bytes.NewBuffer([]byte{})
	table := tablewriter.NewWriter(b)

	table.SetHeader([]string{"Item", "Amount"})

	for k, v := range amounts {
		table.Append([]string{k, fmt.Sprintf("%.2f", v)})
	}

	table.SetRowLine(true)
	table.SetRowSeparator("-")
	table.SetAutoFormatHeaders(false)

	table.Render()

	return fmt.Sprintf("```%s```", b.String())
}
