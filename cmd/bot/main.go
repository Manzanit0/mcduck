package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/cmd/bot/internal/bot"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/pkg/invx"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/trace"
)

const (
	serviceName     = "tgram-bot"
	defaultCurrency = "€"
)

func main() {
	tp, err := initTracerProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			log.Printf("shutdown tracer: %s\n", err.Error())
		}
	}()

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	invxClient := invx.NewClient(os.Getenv("INVX_HOST"), os.Getenv("INVX_AUTH_TOKEN"))
	tgramClient := tgram.NewClient(http.DefaultClient, os.Getenv("TELEGRAM_BOT_TOKEN"))
	mcduckClient := client.NewMcDuckClient(os.Getenv("MCDUCK_HOST"), os.Getenv("MCDUCK_AUTH_TOKEN"))
	r.POST("/telegram/webhook", telegramWebhookController(tgramClient, invxClient, mcduckClient))

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

func telegramWebhookController(tgramClient tgram.Client, invxClient invx.Client, mcduck client.McDuckClient) func(c *gin.Context) {
	return func(c *gin.Context) {
		var r tgram.WebhookRequest
		if err := c.ShouldBindJSON(&r); err != nil {
			log.Println("[ERROR] bind Telegram webhook payload:", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Errorf("payload does not conform with telegram contract: %w", err).Error(),
			})
			return
		}

		switch {
		case r.Message != nil && r.Message.Text != nil && strings.HasPrefix(*r.Message.Text, "/login"):
			c.JSON(http.StatusOK, bot.LoginLink(&r))

		case r.Message == nil || len(r.Message.Photos) > 0:
			c.JSON(http.StatusOK, bot.ParseReceipt(c.Request.Context(), tgramClient, invxClient, mcduck, &r))

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
