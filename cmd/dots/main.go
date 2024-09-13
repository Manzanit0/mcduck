package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"connectrpc.com/otelconnect"
	"github.com/rs/cors"

	"github.com/manzanit0/mcduck/api/auth.v1/authv1connect"
	"github.com/manzanit0/mcduck/api/receipts.v1/receiptsv1connect"
	"github.com/manzanit0/mcduck/cmd/dots/servers"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/micro"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xhttp"
	"github.com/manzanit0/mcduck/pkg/xlog"
	"github.com/manzanit0/mcduck/pkg/xsql"
	"github.com/manzanit0/mcduck/pkg/xtrace"
)

const serviceName = "dots"

func main() {
	if err := run(); err != nil {
		slog.Error("exiting server", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	xlog.InitSlog()

	tp, err := xtrace.TracerFromEnv(context.Background(), serviceName)
	if err != nil {
		return err
	}
	defer tp.Shutdown(context.Background())

	dbx, err := xsql.OpenFromEnv()
	if err != nil {
		return err
	}
	defer xsql.Close(dbx)

	tgramToken := micro.MustGetEnv("TELEGRAM_BOT_TOKEN")
	tgramClient := tgram.NewClient(xhttp.NewClient(), tgramToken)

	parserHost := micro.MustGetEnv("PARSER_HOST")
	parserClient := client.NewParserClient(parserHost)

	otelInterceptor, err := otelconnect.NewInterceptor(otelconnect.WithTrustRemote(), otelconnect.WithoutMetrics())
	if err != nil {
		return err
	}

	authInterceptor := auth.AuthenticationInterceptor()
	traceEnhancer := xtrace.SpanEnhancerInterceptor()

	mux := http.NewServeMux()
	mux.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "pong"}`))
	}))

	mux.Handle(authv1connect.NewAuthServiceHandler(
		servers.NewAuthServer(dbx, tgramClient),
		connect.WithInterceptors(otelInterceptor, traceEnhancer),
	))

	mux.Handle(receiptsv1connect.NewReceiptsServiceHandler(
		servers.NewReceiptsServer(dbx, parserClient, tgramClient),
		connect.WithInterceptors(otelInterceptor, authInterceptor, traceEnhancer),
	))

	return micro.RunGracefully(withCORS(mux))
}

// withCORS adds CORS support to a Connect HTTP handler.
func withCORS(h http.Handler) http.Handler {
	allowedOrigins := micro.MustGetEnv("ALLOWED_ORIGINS")
	slog.Info("allowed origins: " + allowedOrigins)

	middleware := cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigins},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})
	return middleware.Handler(h)
}
