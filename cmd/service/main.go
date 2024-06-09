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
	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	// Used for healthcheck
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Auto-instruments every endpoint
	r.Use(tp.TraceRequests())

	r.GET("/", api.LandingPage)
	r.Use(auth.CookieMiddleware)

	expenseRepository := expense.NewRepository(dbx)

	loggedIn := r.Group("/").Use(api.ForceLogin)
	loggedInAndAuthorised := r.Group("/").Use(api.ForceLogin, api.ExpenseOwnershipWall(expenseRepository))

	controller := api.RegistrationController{DB: dbx.GetSQLX(), Telegram: tgramClient}
	r.GET("/register", controller.GetRegisterForm)
	r.POST("/register", controller.RegisterUser)
	r.GET("/login", controller.GetLoginForm)
	r.POST("/login", controller.LoginUser)
	r.GET("/signout", controller.Signout)
	r.GET("/connect", controller.GetConnectForm)
	r.POST("/connect", controller.ConnectUser)

	data, err := readSampleData()
	if err != nil {
		log.Fatalf("read sample data: %s", err.Error())
	}

	dashController := api.DashboardController{Expenses: expenseRepository, SampleData: data}
	r.GET("/live_demo", dashController.LiveDemo)
	r.POST("/upload", dashController.UploadExpenses)
	loggedIn.GET("/dashboard", dashController.Dashboard)

	expensesController := api.ExpensesController{Expenses: expenseRepository}
	loggedIn.GET("/expenses", expensesController.ListExpenses)
	loggedIn.PUT("/expenses", expensesController.CreateExpense)
	loggedInAndAuthorised.PATCH("/expenses/:id", expensesController.UpdateExpense)
	loggedInAndAuthorised.DELETE("/expenses/:id", expensesController.DeleteExpense)

	// TODO: find a better authwall story, or duplicate the expenses one?
	invxClient := invx.NewClient(os.Getenv("INVX_HOST"), os.Getenv("INVX_AUTH_TOKEN"))
	receiptsController := api.ReceiptsController{Receipts: receipt.NewRepository(dbx), Invx: invxClient}
	r.POST("/receipts", receiptsController.CreateReceipt)
	r.GET("/receipts", receiptsController.ListReceipts)
	r.PATCH("/receipts/:id", receiptsController.UpdateReceipt)

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
