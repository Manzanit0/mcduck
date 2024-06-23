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
	"github.com/go-slog/otelslog"
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
	var handler slog.Handler
	handler = slog.NewTextHandler(os.Stdout, nil) // logfmt
	handler = otelslog.NewHandler(handler)
	handler = xlog.NewDefaultContextHandler(handler)

	logger := slog.New(handler)
	logger = logger.With("service", serviceName)
	slog.SetDefault(logger)

	tp, err := initTracerProvider()
	if err != nil {
		slog.Error("init tracer", "error", err.Error())
		os.Exit(1)
	}

	defer func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			slog.Error("fail to shutdown tracer", "error", err.Error())
		}
	}()

	r := gin.New()
	r.Use(xlog.EnhanceContext)

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

		case r.Message == nil || len(r.Message.Photos) > 0:
			c.JSON(http.StatusOK, bot.ParseReceipt(c.Request.Context(), tgramClient, mcduck, &r))

		default:
			c.JSON(http.StatusOK, tgram.NewMarkdownResponse("Hey\\! Just send me a picture with a receipt, and I'll do the rest\\!", r.GetFromID()))
		}
	}
}

func initTracerProvider() (*trace.Provider, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if endpoint == "" || headers == "" {
		return nil, fmt.Errorf("missing OTEL_EXPORTER_* environment variables")
	}

	opts := trace.NewExporterOptions(endpoint, headers)
	tp, err := trace.InitTracer(context.Background(), serviceName, opts)
	if err != nil {
		return nil, fmt.Errorf("init tracer: %s", err.Error())
	}

	return tp, nil
}
