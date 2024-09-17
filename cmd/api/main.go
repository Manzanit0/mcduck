package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/manzanit0/mcduck/api/auth.v1/authv1connect"
	"github.com/manzanit0/mcduck/api/receipts.v1/receiptsv1connect"
	"github.com/manzanit0/mcduck/cmd/api/controllers"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/micro"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"github.com/manzanit0/mcduck/pkg/xhttp"
	"github.com/manzanit0/mcduck/pkg/xsql"
)

const serviceName = "mcduck"

//go:embed templates/*.html
var templates embed.FS

//go:embed assets/*.ico assets/*.css
var assets embed.FS

//go:embed sample_data.csv
var sampleData embed.FS

func main() {
	if err := run(); err != nil {
		slog.Error("exiting server", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	svc, err := micro.NewGinService(serviceName)
	if err != nil {
		return fmt.Errorf("new gin service: %w", err)
	}

	db, err := xsql.OpenFromEnv()
	if err != nil {
		return fmt.Errorf("open database with tracing: %w", err)
	}
	defer xsql.Close(db)

	tgramToken := micro.MustGetEnv("TELEGRAM_BOT_TOKEN") // TODO: shouldn't throw.
	tgramClient := tgram.NewClient(xhttp.NewClient(), tgramToken)

	t, err := template.ParseFS(templates, "templates/*.html")
	if err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}

	r := svc.Engine
	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	authClient := authv1connect.NewAuthServiceClient(xhttp.NewClient(), micro.MustGetEnv("PRIVATE_DOTS_HOST"))
	registrationController := controllers.RegistrationController{
		DB:              db,
		Telegram:        tgramClient,
		AuthServiceHost: micro.MustGetEnv("PUBLIC_DOTS_HOST"),
		AuthClient:      authClient,
	}

	expenseRepository := expense.NewRepository(db)
	expensesController := controllers.ExpensesController{Expenses: expenseRepository}

	receiptsClient := receiptsv1connect.NewReceiptsServiceClient(xhttp.NewClient(), micro.MustGetEnv("PRIVATE_DOTS_HOST"))
	parserHost := micro.MustGetEnv("PARSER_HOST") // TODO: shouldn't throw.
	parserClient := client.NewParserClient(parserHost)
	receiptsRepository := receipt.NewRepository(db)
	receiptsController := controllers.ReceiptsController{
		Expenses:       expenseRepository,
		Receipts:       receiptsRepository,
		Parser:         parserClient,
		ReceiptsClient: receiptsClient,
	}

	data, err := readSampleData()
	if err != nil {
		return fmt.Errorf("read sample data: %w", err)
	}
	dashController := controllers.DashboardController{Expenses: expenseRepository, SampleData: data}

	nologin := r.
		Group("/").
		Use(auth.CookieMiddleware)

	nologin.GET("/", controllers.LandingPage)
	nologin.GET("/register", registrationController.GetRegisterForm)
	nologin.GET("/login", registrationController.GetLoginForm)
	nologin.GET("/signout", registrationController.Signout)
	nologin.GET("/connect", registrationController.GetConnectForm)
	nologin.GET("/live_demo", dashController.LiveDemo)
	nologin.POST("/upload", dashController.UploadExpenses)

	loggedIn := r.
		Group("/").
		Use(auth.CookieMiddleware).
		Use(controllers.ForceLogin)

	loggedIn.GET("/dashboard", dashController.Dashboard)
	loggedIn.GET("/receipts", receiptsController.ListReceipts)
	loggedIn.GET("/receipts/:id/review", receiptsController.ReviewReceipt)
	loggedIn.GET("/expenses", expensesController.ListExpenses)

	apiG := r.
		Group("/").
		Use(auth.CookieMiddleware). // Add cookie auth so the frontend can talk easily with the backend.
		Use(auth.BearerMiddleware).
		Use(controllers.ForceAuthentication)

	ownsReceipt := r.
		Group("/").
		Use(auth.CookieMiddleware).
		Use(auth.BearerMiddleware).
		Use(controllers.ForceAuthentication).
		Use(controllers.ReceiptOwnershipWall(receiptsRepository))

	ownsReceipt.PATCH("/receipts/:id", receiptsController.UpdateReceipt)
	ownsReceipt.DELETE("/receipts/:id", receiptsController.DeleteReceipt)
	ownsReceipt.GET("/receipts/:id/image", receiptsController.GetImage)
	apiG.POST("/receipts/upload", receiptsController.UploadReceipts)

	ownsExpense := r.
		Group("/").
		Use(auth.CookieMiddleware).
		Use(auth.BearerMiddleware).
		Use(controllers.ForceAuthentication).
		Use(controllers.ExpenseOwnershipWall(expenseRepository))

	ownsExpense.PATCH("/expenses/:id", expensesController.UpdateExpense)
	ownsExpense.DELETE("/expenses/:id", expensesController.DeleteExpense)
	apiG.PUT("/expenses", expensesController.CreateExpense)
	apiG.POST("/expenses/merge", expensesController.MergeExpenses) // TODO: this should be under receipts with authz?

	return svc.Run()
}

func readSampleData() ([]expense.Expense, error) {
	b, err := sampleData.ReadFile("sample_data.csv")
	if err != nil {
		return nil, err
	}

	return expense.FromCSV(bytes.NewReader(b))
}
