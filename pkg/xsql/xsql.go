package xsql

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/manzanit0/isqlx"
	"go.opentelemetry.io/otel"
)

func Open(serviceName string) (isqlx.DBX, error) {
	tracer := otel.Tracer(serviceName)

	config, err := configFromEnv()
	if err != nil {
		return nil, fmt.Errorf("read config from environment variables: %w", err)
	}

	dbx, err := isqlx.NewDBXFromConfig("pgx", &isqlx.DBConfig{
		Host:     config.host,
		Port:     config.port,
		User:     config.user,
		Password: config.password,
		Name:     config.name,
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

func Close(db *sqlx.DB) {
	err := db.Close()
	if err != nil {
		slog.Error("fail to close database connection", "error", err.Error())
	}
}

func configFromEnv() (*config, error) {
	portStr := os.Getenv("PGPORT")
	if portStr == "" {
		return nil, fmt.Errorf("PGPORT is empty")
	}

	portInt, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("parse db port from env var PGPORT: %w", err)
	}

	host := os.Getenv("PGHOST")
	if host == "" {
		return nil, fmt.Errorf("PGHOST is empty")
	}

	user := os.Getenv("PGUSER")
	if user == "" {
		return nil, fmt.Errorf("PGUSER is empty")
	}

	password := os.Getenv("PGPASSWORD")
	if password == "" {
		return nil, fmt.Errorf("PGPASSWORD is empty")
	}

	name := os.Getenv("PGDATABASE")
	if name == "" {
		return nil, fmt.Errorf("PGDATABASE is empty")
	}

	c := config{
		host:     host,
		port:     portInt,
		user:     user,
		password: password,
		name:     name,
	}

	return &c, nil
}

type config struct {
	host     string
	port     int
	user     string
	password string
	name     string
}
