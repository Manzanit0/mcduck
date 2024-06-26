package micro

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/manzanit0/mcduck/pkg/trace"
	"github.com/manzanit0/mcduck/pkg/xlog"
)

type Service struct {
	Name   string
	Engine *gin.Engine
	tp     *trace.Provider
}

func NewGinService(name string) (Service, error) {
	xlog.InitSlog()

	tp, err := trace.TracerFromEnv(context.Background(), name)
	if err != nil {
		return Service{}, fmt.Errorf("get tracer from env %w", err)
	}

	r := gin.Default()
	r.Use(xlog.EnhanceContext)
	r.Use(tp.TraceRequests())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	return Service{Name: name, Engine: r}, nil
}

func (s *Service) Run() error {
	defer func() {
		err := s.tp.Shutdown(context.Background())
		if err != nil {
			slog.Error("fail to shutdown tracer", "error", err.Error())
		}
	}()

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: s.Engine}
	go func() {
		slog.Info(fmt.Sprintf("serving HTTP on :%s", port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server ended abruptly: %s", "error", err.Error())
		} else {
			slog.Info("server ended gracefully")
		}

		stop()
	}()

	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server exited")
	return nil
}

func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		slog.Error(fmt.Sprintf("environment variable %s is empty", key))
		os.Exit(1)
	}

	return value
}
