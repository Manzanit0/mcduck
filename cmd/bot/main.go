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
	"github.com/manzanit0/mcduck/pkg/xtrace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
		ctx := c.Request.Context()

		span := trace.SpanFromContext(ctx)

		var r tgram.WebhookRequest
		if err := c.ShouldBindJSON(&r); err != nil {
			xtrace.RecordError(ctx, "unable to bind request payload", err)

			// FIXME: this actually isn't what's happening. It's not a json.Unmarshall as I expected.
			res := gin.H{"error": fmt.Sprintf("payload does not conform with telegram contract: %s", err.Error())}
			c.JSON(http.StatusBadRequest, res)
			return
		}

		span.SetAttributes(
			attribute.Int("mduck.telegram.chat_id", r.GetFromID()),
			attribute.String("mduck.telegram.language_code", r.GetFromLanguageCode()),
		)

		switch {
		case r.Message != nil && r.Message.Text != nil && strings.HasPrefix(*r.Message.Text, "/login"):
			span.SetAttributes(attribute.String("mduck.telegram.command", "login"))

			res := bot.LoginLink(ctx, &r)
			c.JSON(http.StatusOK, res)

			// The message has either photos or a doc.
		case r.Message != nil && (len(r.Message.Photos) > 0 || r.Message.Document != nil):
			span.SetAttributes(attribute.String("mduck.telegram.command", "upload"))

			res := bot.ParseReceipt(ctx, tgramClient, mcduck, &r)
			c.JSON(http.StatusOK, res)

		default:
			span.SetAttributes(attribute.String("mduck.telegram.command", "unknown"))

			res := tgram.NewMarkdownResponse("Hey\\! Just send me a picture with a receipt and I will do the rest\\!", r.GetFromID())
			c.JSON(http.StatusOK, res)
		}
	}
}
