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

	r := gin.Default()
	r.SetHTMLTemplate(t)
	r.StaticFS("/public", http.FS(assets))

	db, err := sql.Open("pgx", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv("PGUSER"), os.Getenv("PGPASSWORD"), os.Getenv("PGHOST"), os.Getenv("PGPORT"), os.Getenv("PGDATABASE")))
	if err != nil {
		log.Fatalf("unable to open db conn: %s", err.Error())
	}

	defer func() {
		err = db.Close()
		if err != nil {
			log.Printf("error closing db connection: %s\n", err.Error())
		}
	}()

	r.GET("/", api.LandingPage)

	r.Use(auth.CookieMiddleware)

	controller := api.RegistrationController{DB: sqlx.NewDb(db, "postgres")}
	r.GET("/register", controller.GetRegisterForm)
	r.POST("/register", controller.RegisterUser)
	r.GET("/login", controller.GetLoginForm)
	r.POST("/login", controller.LoginUser)
	r.GET("/signout", controller.Signout)

	data, err := readSampleData()
	if err != nil {
		log.Fatal(err)
	}

	dashController := api.DashboardController{DB: sqlx.NewDb(db, "postgres"), SampleData: data}
	r.GET("/live_demo", dashController.LiveDemo)
	r.POST("/upload", dashController.UploadExpenses)
	r.GET("/dashboard", api.AuthWall(dashController.Dashboard))

	expensesController := api.ExpensesController{DB: sqlx.NewDb(db, "postgres")}
	r.GET("/expenses", api.AuthWall(expensesController.ListExpenses))

	var port string
	if port = os.Getenv("PORT"); port == "" {
		port = "8080"
	}

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal(err)
	}
}

func readSampleData() ([]expense.Expense, error) {
	b, err := sampleData.ReadFile("sample_data.csv")
	if err != nil {
		return nil, err
	}

	return expense.FromCSV(bytes.NewReader(b))
}
