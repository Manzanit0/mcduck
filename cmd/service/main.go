package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/manzanit0/isqlx"
	"go.opentelemetry.io/otel"

	"github.com/manzanit0/mcduck/cmd/service/api"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/invx"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/trace"
)

const serviceName = "mcduck"

//go:embed templates/*.html
var templates embed.FS

//go:embed assets/*.ico assets/*.css
var assets embed.FS

//go:embed sample_data.csv
var sampleData embed.FS

func main() {
	tp, err := trace.TracerFromEnv(context.Background(), serviceName)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			log.Printf("shutdown tracer: %s\n", err.Error())
		}
	}()

	dbx, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = dbx.GetSQLX().Close()
		if err != nil {
			log.Printf("closing postgres connection: %s\n", err.Error())
		}
	}()

	tgramClient := tgram.NewClient(http.DefaultClient, os.Getenv("TELEGRAM_BOT_TOKEN"))

	t, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	// Auto-instruments every endpoint
	r.Use(tp.TraceRequests())

	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	// Used for healthcheck
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	registrationController := api.RegistrationController{DB: dbx.GetSQLX(), Telegram: tgramClient}

	expenseRepository := expense.NewRepository(dbx)
	expensesController := api.ExpensesController{Expenses: expenseRepository}

	invxClient := invx.NewClient(os.Getenv("INVX_HOST"), os.Getenv("INVX_AUTH_TOKEN"))
	receiptsRepository := receipt.NewRepository(dbx)
	receiptsController := api.ReceiptsController{Receipts: receiptsRepository, Invx: invxClient, Expenses: expenseRepository}

	data, err := readSampleData()
	if err != nil {
		log.Fatalf("read sample data: %s", err.Error())
	}
	dashController := api.DashboardController{Expenses: expenseRepository, SampleData: data}

	nologin := r.
		Group("/").
		Use(auth.CookieMiddleware)

	nologin.GET("/", api.LandingPage)
	nologin.GET("/register", registrationController.GetRegisterForm)
	nologin.POST("/register", registrationController.RegisterUser)
	nologin.GET("/login", registrationController.GetLoginForm)
	nologin.POST("/login", registrationController.LoginUser)
	nologin.GET("/signout", registrationController.Signout)
	nologin.GET("/connect", registrationController.GetConnectForm)
	nologin.POST("/connect", registrationController.ConnectUser)
	nologin.GET("/live_demo", dashController.LiveDemo)
	nologin.POST("/upload", dashController.UploadExpenses)

	loggedIn := r.
		Group("/").
		Use(auth.CookieMiddleware).
		Use(api.ForceLogin)

	loggedIn.GET("/dashboard", dashController.Dashboard)
	loggedIn.GET("/receipts", receiptsController.ListReceipts)
	loggedIn.GET("/receipts/:id/review", receiptsController.ReviewReceipt)
	loggedIn.GET("/expenses", expensesController.ListExpenses)

	apiG := r.
		Group("/").
		Use(auth.CookieMiddleware). // Add cookie auth so the frontend can talk easily with the backend.
		Use(auth.BearerMiddleware).
		Use(api.ForceAuthentication)

		// This is a quick hack to get around the fact that the API doesn't support
		// PATs or similar mechanics.
	nologin.POST("/x/login", registrationController.LoginAPI)

	ownsReceipt := r.
		Group("/").
		Use(auth.CookieMiddleware).
		Use(auth.BearerMiddleware).
		Use(api.ForceAuthentication).
		Use(api.ReceiptOwnershipWall(receiptsRepository))

	ownsReceipt.PATCH("/receipts/:id", receiptsController.UpdateReceipt)
	ownsReceipt.DELETE("/receipts/:id", receiptsController.DeleteReceipt)
	apiG.POST("/receipts", receiptsController.CreateReceipt) // TODO: this should be a PUT?

	ownsExpense := r.
		Group("/").
		Use(auth.CookieMiddleware).
		Use(auth.BearerMiddleware).
		Use(api.ForceAuthentication).
		Use(api.ExpenseOwnershipWall(expenseRepository))

	ownsExpense.PATCH("/expenses/:id", expensesController.UpdateExpense)
	ownsExpense.DELETE("/expenses/:id", expensesController.DeleteExpense)
	apiG.PUT("/expenses", expensesController.CreateExpense)
	apiG.POST("/expenses/merge", expensesController.MergeExpenses) // TODO: this should be under receipts with authz?

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

func openDB() (isqlx.DBX, error) {
	tracer := otel.Tracer(serviceName)
	port, err := strconv.Atoi(os.Getenv("PGPORT"))
	if err != nil {
		return nil, fmt.Errorf("parse db port from env var PGPORT: %w", err)
	}

	dbx, err := isqlx.NewDBXFromConfig("pgx", &isqlx.DBConfig{
		Host:     os.Getenv("PGHOST"),
		Port:     port,
		User:     os.Getenv("PGUSER"),
		Password: os.Getenv("PGPASSWORD"),
		Name:     os.Getenv("PGDATABASE"),
	}, tracer)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	err = dbx.GetSQLX().DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	return dbx, nil
}

func readSampleData() ([]expense.Expense, error) {
	b, err := sampleData.ReadFile("sample_data.csv")
	if err != nil {
		return nil, err
	}

	return expense.FromCSV(bytes.NewReader(b))
}
