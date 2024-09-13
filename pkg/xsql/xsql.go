package xsql

import (
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func OpenFromEnv() (*sqlx.DB, error) {
	driverName, err := otelsql.Register("pgx", otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, fmt.Errorf("register otelsql: %w", err)
	}

	config, err := configFromEnv()
	if err != nil {
		return nil, fmt.Errorf("read config from environment variables: %w", err)
	}

	db, err := sqlx.Open(driverName, config.url())
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	return db, nil
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

func (c *config) url() string {
	query := url.Values{}
	query.Set("client_encoding", "UTF8")

	datasource := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.user, c.password),
		Host:     net.JoinHostPort(c.host, strconv.Itoa(c.port)),
		Path:     c.name,
		RawQuery: query.Encode(),
	}

	return datasource.String()
}

func TxClose(tx *sqlx.Tx) {
	if r := recover(); r != nil {
		slog.Error("recovered an error in TxClose()", "error", r)
		_ = tx.Rollback()
		panic(r)
	} else {
		// Transaction leak failsafe:
		//
		// I don't check the errors here because the transaction might already
		// have been committed/rolledback. If there's an issue with the database
		// connection we'll catch it the next time that db handle gets used.
		_ = tx.Rollback()
	}
}
