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
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	otrace "go.opentelemetry.io/otel/trace"

	"github.com/manzanit0/isqlx"
	"github.com/manzanit0/mcduck/cmd/service/api"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/expense"
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
	t, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if endpoint == "" || headers == "" {
		log.Fatal("missing OTEL_EXPORTER_* environment variables")
	}

	opts := trace.NewExporterOptions(endpoint, headers)
	tp, err := trace.InitTracer(context.Background(), serviceName, opts)
	if err != nil {
		log.Fatalf("init tracer: %s", err.Error())
	}

	defer func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			log.Fatalf("shutdown tracer: %s", err.Error())
		}
	}()

	// Auto-instruments every endpoint
	r.Use(otelgin.Middleware(serviceName))

	tracer := otel.Tracer(serviceName)
	dbx := MustOpenDB(tracer)
	defer func() {
		err = dbx.GetSQLX().Close()
		if err != nil {
			log.Printf("closing postgres connection: %s\n", err.Error())
		}
	}()

	r.GET("/", api.LandingPage)
	r.Use(auth.CookieMiddleware)

	expenseRepository := expense.NewRepository(dbx)

	loggedIn := r.Group("/").Use(api.ForceLogin)
	loggedInAndAuthorised := r.Group("/").Use(api.ForceLogin, api.ExpenseOwnershipWall(expenseRepository))

	controller := api.RegistrationController{DB: dbx.GetSQLX()}
	r.GET("/register", controller.GetRegisterForm)
	r.POST("/register", controller.RegisterUser)
	r.GET("/login", controller.GetLoginForm)
	r.POST("/login", controller.LoginUser)
	r.GET("/signout", controller.Signout)

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

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}

func MustOpenDB(tracer otrace.Tracer) isqlx.DBX {
	port, err := strconv.Atoi(os.Getenv("PGPORT"))
	if err != nil {
		log.Fatalf("parse db port from env var PGPORT: %s", err.Error())
	}

	dbx, err := isqlx.NewDBXFromConfig("pgx", &isqlx.DBConfig{
		Host:     os.Getenv("PGHOST"),
		Port:     port,
		User:     os.Getenv("PGUSER"),
		Password: os.Getenv("PGPASSWORD"),
		Name:     os.Getenv("PGDATABASE"),
	}, tracer)
	if err != nil {
		log.Fatalf("open postgres connection: %s", err.Error())
	}

	err = dbx.GetSQLX().DB.Ping()
	if err != nil {
		log.Fatalf("ping postgres connection: %s", err.Error())
	}

	return dbx
}

func readSampleData() ([]expense.Expense, error) {
	b, err := sampleData.ReadFile("sample_data.csv")
	if err != nil {
		return nil, err
	}

	return expense.FromCSV(bytes.NewReader(b))
}
