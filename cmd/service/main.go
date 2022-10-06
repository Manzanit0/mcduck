package main

import (
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/manzanit0/mcduck/cmd/service/api"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/expense"
)

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

	dbx := MustOpenDB()
	defer func() {
		err = dbx.Close()
		if err != nil {
			log.Printf("closing postgres connection: %s\n", err.Error())
		}
	}()

	data, err := readSampleData()
	if err != nil {
		log.Fatalf("read sample data: %s", err.Error())
	}

	r := gin.Default()
	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	r.GET("/", api.LandingPage)
	r.Use(auth.CookieMiddleware)

	authenticated := r.Group("/").Use(api.AuthWall)
	authorised := r.Group("/").Use(api.AuthWall, api.ExpenseOwnershipWall(dbx))

	controller := api.RegistrationController{DB: dbx}
	r.GET("/register", controller.GetRegisterForm)
	r.POST("/register", controller.RegisterUser)
	r.GET("/login", controller.GetLoginForm)
	r.POST("/login", controller.LoginUser)
	r.GET("/signout", controller.Signout)

	dashController := api.DashboardController{DB: dbx, SampleData: data}
	r.GET("/live_demo", dashController.LiveDemo)
	r.POST("/upload", dashController.UploadExpenses)
	authenticated.GET("/dashboard", dashController.Dashboard)

	expensesController := api.ExpensesController{DB: dbx}
	authenticated.GET("/expenses", expensesController.ListExpenses)
	authenticated.PUT("/expenses", expensesController.CreateExpense)
	authorised.PATCH("/expenses/:id", expensesController.UpdateExpense)
	authorised.DELETE("/expenses/:id", expensesController.DeleteExpense)

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}

func MustOpenDB() *sqlx.DB {
	db, err := sql.Open("pgx", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGDATABASE")))
	if err != nil {
		log.Fatalf("open postgres connection: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("ping postgres connection: %s", err.Error())
	}

	return sqlx.NewDb(db, "postgres")
}

func readSampleData() ([]expense.Expense, error) {
	b, err := sampleData.ReadFile("sample_data.csv")
	if err != nil {
		return nil, err
	}

	return expense.FromCSV(bytes.NewReader(b))
}
