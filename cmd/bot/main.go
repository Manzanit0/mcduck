package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/cmd/bot/internal/bot"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/pkg/micro"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xhttp"
)

const (
	serviceName = "tgram-bot"
)

func main() {
	svc, err := micro.NewGinService(serviceName)
	if err != nil {
		panic(err)
	}

	tgramToken := micro.MustGetEnv("TELEGRAM_BOT_TOKEN")
	mcduckHost := micro.MustGetEnv("MCDUCK_HOST")

	h := xhttp.NewClient()
	tgramClient := tgram.NewClient(h, tgramToken)
	mcduckClient := client.NewMcDuckClient(mcduckHost)

	svc.Engine.POST("/telegram/webhook", telegramWebhookController(tgramClient, mcduckClient))

	if err := svc.Run(); err != nil {
		slog.Error("run ended with error", "error", err.Error())
		os.Exit(1)
	}
}

func telegramWebhookController(tgramClient tgram.Client, mcduck client.McDuckClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		var r tgram.WebhookRequest
		if err := c.ShouldBindJSON(&r); err != nil {
			slog.ErrorContext(c.Request.Context(), "bind Telegram webhook payload", "error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Errorf("payload does not conform with telegram contract: %w", err).Error(),
			})
			return
		}

		switch {
		case r.Message != nil && r.Message.Text != nil && strings.HasPrefix(*r.Message.Text, "/login"):
			c.JSON(http.StatusOK, bot.LoginLink(c.Request.Context(), &r))

			// The message has either photos or a doc.
		case r.Message != nil && (len(r.Message.Photos) > 0 || r.Message.Document != nil):
			c.JSON(http.StatusOK, bot.ParseReceipt(c.Request.Context(), tgramClient, mcduck, &r))

		default:
			c.JSON(http.StatusOK, tgram.NewMarkdownResponse("Hey\\! Just send me a picture with a receipt, and I'll do the rest\\!", r.GetFromID()))
		}
	}
}
