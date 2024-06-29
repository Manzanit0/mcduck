package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/cmd/bot/internal/bot"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/trace"
	"github.com/manzanit0/mcduck/pkg/xlog"
)

const (
	serviceName     = "tgram-bot"
	defaultCurrency = "€"
)

func main() {
	xlog.InitSlog()

	tp, err := trace.TracerFromEnv(context.Background(), serviceName)
	if err != nil {
		slog.Error("get tracer from env", "error", err.Error())
		os.Exit(1)
	}

	defer func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			slog.Error("fail to shutdown tracer", "error", err.Error())
		}
	}()

	r := gin.Default()
	r.Use(xlog.EnhanceContext)
	r.Use(tp.TraceRequests())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	tgramClient := tgram.NewClient(http.DefaultClient, os.Getenv("TELEGRAM_BOT_TOKEN"))
	mcduckClient := client.NewMcDuckClient(os.Getenv("MCDUCK_HOST"))
	r.POST("/telegram/webhook", telegramWebhookController(tgramClient, mcduckClient))

	// background job to ping users on weather changes
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: r}
	go func() {
		slog.Info("serving HTTP on :%s" + port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server ended abruptly: %s", "error", err.Error())
		} else {
			slog.Info("server ended gracefully")
		}

		stop()
	}()

	// Listen for OS interrupt
	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server exited")
}

func telegramWebhookController(tgramClient tgram.Client, mcduck client.McDuckClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		var r tgram.WebhookRequest
		if err := c.ShouldBindJSON(&r); err != nil {
			slog.Error("bind Telegram webhook payload", "error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Errorf("payload does not conform with telegram contract: %w", err).Error(),
			})
			return
		}

		switch {
		case r.Message != nil && r.Message.Text != nil && strings.HasPrefix(*r.Message.Text, "/login"):
			c.JSON(http.StatusOK, bot.LoginLink(&r))

		case r.Message != nil && len(r.Message.Photos) > 0:
			c.JSON(http.StatusOK, bot.ParseReceipt(c.Request.Context(), tgramClient, mcduck, &r))

		default:
			c.JSON(http.StatusOK, tgram.NewMarkdownResponse("Hey\\! Just send me a picture with a receipt, and I'll do the rest\\!", r.GetFromID()))
		}
	}
}
