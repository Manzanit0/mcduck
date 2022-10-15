package isqlx

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ExecFunc = func(ctx context.Context, query string, arg ...interface{}) (sql.Result, error)
type GetFunc = func(ctx context.Context, dest interface{}, query string, args ...interface{}) error
type SelectFunc = func(ctx context.Context, dest interface{}, query string, args ...interface{}) error

// Querier is an interface which exposes functions to run queries. It's a subset
// of sqlx library functions.
type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// DBX is an interface to make single queries without leveraging transactions.
type DBX interface {
	Querier
	Begin(ctx context.Context) (TX, error)

	// GetSQLX is a way to escape the abstraction when needed.
	GetSQLX() *sqlx.DB
}

// TX is an interface to make queries within a database transaction.  Every
// transaction should invoke TxClose() deferred to make sure that there aren't
// any transaction leaks and that they are rolledback in case of error or panic,
// and Commit() to commit the transaction.
type TX interface {
	Querier
	Commit(ctx context.Context) error
	TxClose(ctx context.Context)
}

type DBConfig struct {
	Host                          string
	Port                          int
	User                          string
	Password                      string
	Name                          string
	MaxConnections                int
	MaxIdleConnections            int
	ConnectionLifetimeSeconds     time.Duration
	IdleConnectionLifetimeSeconds time.Duration
}

func (c *DBConfig) DSN(driver string) (string, error) {
	switch driver {
	case "pgx":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
			c.User,
			c.Password,
			c.Host,
			c.Port,
			c.Name,
		), nil
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			c.User,
			c.Password,
			c.Host,
			c.Port,
			c.Name,
		), nil
	}

	return "", fmt.Errorf("unsupported driver, currently only pgx and mysql are supported")
}

const (
	DefaultMaxConnections                = 20
	DefaultMaxIdleConnections            = 10
	DefaultConnectionLifetimeSeconds     = time.Duration(180) // default to 3 minutes
	DefaultIdleConnectionLifeTimeSeconds = time.Duration(60)  // default to 1 minute
)

func NewDBXFromConfig(driver string, config *DBConfig, tracer trace.Tracer) (DBX, error) {
	dsn, err := config.DSN(driver)
	if err != nil {
		return nil, fmt.Errorf("building DSN for connection: %w", err)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	maxConnections := DefaultMaxConnections
	if config.MaxConnections != 0 {
		maxConnections = config.MaxConnections
	}

	maxIdleConnections := DefaultMaxIdleConnections
	if config.MaxIdleConnections != 0 {
		maxIdleConnections = config.MaxIdleConnections
	}

	connectionLifetimeSeconds := DefaultConnectionLifetimeSeconds
	if config.ConnectionLifetimeSeconds != 0 {
		connectionLifetimeSeconds = config.ConnectionLifetimeSeconds
	}

	idleConnectionLifetimeSeconds := DefaultIdleConnectionLifeTimeSeconds
	if config.IdleConnectionLifetimeSeconds != 0 {
		idleConnectionLifetimeSeconds = config.IdleConnectionLifetimeSeconds
	}

	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	db.SetConnMaxLifetime(time.Second * connectionLifetimeSeconds)
	db.SetConnMaxIdleTime(time.Second * idleConnectionLifetimeSeconds)

	d := sqlx.NewDb(db, driver)
	return &dbx{DB: d, driver: driver, config: config, tracer: tracer}, nil
}

type dbx struct {
	DB     *sqlx.DB
	driver string
	tracer trace.Tracer
	config *DBConfig
}

type tx struct {
	db     *sqlx.DB
	TX     *sqlx.Tx
	driver string
	tracer trace.Tracer
	config *DBConfig
}

func (d *dbx) GetSQLX() *sqlx.DB {
	return d.DB
}

func (d *dbx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return getContext(ctx, d.tracer, d.DB.GetContext, d.driver, d.config, d.DB.Stats(), dest, query, args...)
}

func (d *dbx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return selectContext(ctx, d.tracer, d.DB.SelectContext, d.driver, d.config, d.DB.Stats(), dest, query, args...)
}

func (d *dbx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return execContext(ctx, d.tracer, d.DB.ExecContext, d.driver, d.config, d.DB.Stats(), query, args...)
}

func (d *dbx) Begin(_ context.Context) (TX, error) {
	t, err := d.DB.Beginx()
	if err != nil {
		return nil, err
	}

	return &tx{TX: t, db: d.DB, driver: d.driver, tracer: d.tracer, config: d.config}, nil
}

func (t *tx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return getContext(ctx, t.tracer, t.TX.GetContext, t.driver, t.config, t.db.Stats(), dest, query, args...)
}

func (t *tx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return selectContext(ctx, t.tracer, t.TX.SelectContext, t.driver, t.config, t.db.Stats(), dest, query, args...)
}

func (t *tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return execContext(ctx, t.tracer, t.TX.ExecContext, t.driver, t.config, t.db.Stats(), query, args...)
}

func (t *tx) Commit(ctx context.Context) error {
	_, span := newSpan(ctx, t.driver, "commit", t.config, t.db.Stats(), t.tracer)
	defer span.End()

	err := t.TX.Commit()
	if err != nil {
		span.RecordError(err)
	}

	return err
}

// TxClose makes sure the transaction gets rolled back. It should be run within
// a `defer` statement so it can rollback transactions even in the case of
// panics.
func (t *tx) TxClose(ctx context.Context) {
	_, span := newSpan(ctx, t.driver, "rollback", t.config, t.db.Stats(), t.tracer)
	defer span.End()

	if r := recover(); r != nil {
		log.Printf("recovered an error in TxClose(): %#v", r)
		_ = t.TX.Rollback()
		panic(r)
	} else {
		// Transaction leak failsafe:
		//
		// I don't check the errors here because the transaction might already
		// have been committed/rolledback. If there's an issue with the database
		// connection we'll catch it the next time that db handle gets used.
		_ = t.TX.Rollback()
	}
}

func getContext(
	ctx context.Context,
	tracer trace.Tracer,
	getFn GetFunc,
	driver string,
	config *DBConfig,
	stats sql.DBStats,
	dest interface{},
	query string,
	args ...interface{},
) error {
	ctx, span := newSpan(ctx, driver, query, config, stats, tracer)
	defer span.End()

	span.addQueryParams(args)

	err := getFn(ctx, dest, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			span.addAffectedRowsAttribute(0)
		} else {
			span.RecordError(err)
		}
	}

	return err
}

func selectContext(
	ctx context.Context,
	tracer trace.Tracer,
	selectFn SelectFunc,
	driver string,
	config *DBConfig,
	stats sql.DBStats,
	dest interface{},
	query string,
	args ...interface{},
) error {
	ctx, span := newSpan(ctx, driver, query, config, stats, tracer)
	defer span.End()

	span.addQueryParams(args)

	err := selectFn(ctx, dest, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			span.addAffectedRowsAttribute(0)
		} else {
			span.RecordError(err)
		}
	} else if n := getReturnedRows(dest); n != -1 {
		span.addAffectedRowsAttribute(int64(n))
	}

	return err
}

func execContext(
	ctx context.Context,
	tracer trace.Tracer,
	execFn ExecFunc,
	driver string,
	config *DBConfig,
	stats sql.DBStats,
	query string,
	args ...interface{},
) (sql.Result, error) {
	ctx, span := newSpan(ctx, driver, query, config, stats, tracer)
	defer span.End()

	r, err := execFn(ctx, query, args...)
	if err != nil {
		if err != sql.ErrNoRows {
			span.RecordError(err)
		}
	} else if n, err := r.RowsAffected(); err != nil {
		span.addAffectedRowsAttribute(n)
	}

	return r, err
}

func parseQueryOperation(query string) string {
	query = strings.ToLower(query)
	if strings.HasPrefix(query, "update") { // nolint: gocritic
		return "update"
	} else if strings.HasPrefix(query, "select") {
		return "select"
	} else if strings.HasPrefix(query, "insert") {
		return "insert"
	} else if strings.HasPrefix(query, "delete") {
		return "delete"
	} else if strings.HasPrefix(query, "commit") {
		return "commit"
	} else if strings.HasPrefix(query, "rollback") {
		return "rollback"
	}

	return "unknown"
}

// getReturnedRows extracts the amount of rows returned from dest assuming that it's
// the result of a database operation.
// @see https://goplay.tools/snippet/oKaFkTexWBk
func getReturnedRows(dest interface{}) int {
	t := reflect.TypeOf(dest)

	switch t.Kind() {
	case reflect.Slice:
		return reflect.ValueOf(dest).Len()
	case reflect.Array:
		return t.Len()
	default:
		return -1
	}
}

// customSpan is simply a wrapper around trace.Span to provide some commodity
// functions for adding attributes to the trace.
type customSpan struct {
	trace.Span
}

func newSpan(ctx context.Context, driver, query string, config *DBConfig, stats sql.DBStats, tracer trace.Tracer) (context.Context, *customSpan) {
	name := inferSpanName(query, config.Name)
	ctx, span := tracer.Start(ctx, name)
	custom := customSpan{span}

	op := parseQueryOperation(query)
	addDatabaseQueryAttributes(&custom, query, op)
	addDatabaseSystemAttributes(&custom, driver)
	addDatabaseConnectionAttributes(&custom, config.Host, config.Port, config.Name, config.User)
	addDatabaseStatsAttributes(&custom, stats)

	return ctx, &custom
}

func (s *customSpan) addQueryParams(args ...interface{}) {
	for i := range args {
		v := fmt.Sprint(args[i])
		s.addQueryParamAttribute(fmt.Sprint(i), fmt.Sprint(v))
	}
}

func (s *customSpan) addAffectedRowsAttribute(n int64) {
	s.SetAttributes(attribute.Int64("db.result.returned_rows", n))
}

func (s *customSpan) addQueryParamAttribute(k, v string) {
	s.SetAttributes(attribute.String(fmt.Sprintf("db.statement.param_%s", k), v))
}

func addDatabaseSystemAttributes(s *customSpan, driver string) {
	s.SetAttributes(attribute.String("db.system", driver))
}

func addDatabaseConnectionAttributes(s *customSpan, host string, port int, dbName, user string) {
	if s == nil {
		return
	}

	s.SetAttributes(
		attribute.String("net.peer.transport", "IP.TCP"),
		attribute.String("net.peer.name", host),
		attribute.Int("net.peer.port", port),
		attribute.String("db.name", dbName),
		attribute.String("db.user", user),
	)
}

func addDatabaseQueryAttributes(s *customSpan, query, operation string) {
	if s == nil {
		return
	}

	s.SetAttributes(
		attribute.String("db.statement", query),
		attribute.String("db.operation", operation),
	)
}

func addDatabaseStatsAttributes(s *customSpan, stats sql.DBStats) {
	if s == nil {
		return
	}

	s.SetAttributes(
		// Pool config
		attribute.Int("db.sql.max_open_connection", stats.MaxOpenConnections),

		// Pool status
		attribute.Int("db.sql.open_connections", stats.OpenConnections),
		attribute.Int("db.sql.in_use", stats.InUse),
		attribute.Int("db.sql.idle", stats.Idle),

		// Counters
		attribute.Int64("db.sql.wait_count", stats.WaitCount),
		attribute.Int64("db.sql.wait_duration_ms", stats.WaitDuration.Milliseconds()),
		attribute.Int64("db.sql.max_idle_closed", stats.MaxIdleClosed),
		attribute.Int64("db.sql.max_idle_time_closed", stats.MaxIdleTimeClosed),
		attribute.Int64("db.sql.max_life_time_closed", stats.MaxLifetimeClosed),
	)
}

func inferSpanName(statement, database string) string { //nolint: gocyclo
	operation := parseQueryOperation(statement)

	var tableName string
	var awkwardScenario bool

	// Incredibly naive way to extract the table name from the query.  This will
	// work for simple queries, but it will not work for any query that contains
	// funny joins, inner selects, or other complex constructs.
	arr := strings.Split(statement, " ")
	for k, v := range arr {
		// Inserts are always in the form of "INSERT INTO table_name".
		if strings.EqualFold(v, "insert") && len(arr) > k+2 {
			tableName = arr[k+2]
			break
		}

		// Inserts are always in the form of "UPDATE table_name".
		if strings.EqualFold(v, "update") && len(arr) > k+1 {
			tableName = arr[k+1]
			break
		}

		// Deletes are always in the form of "DELETE FROM table_name".
		// It's also relevant for the algorithm to check for deletes before
		// selects because they both contain a "FROM" keyword.
		if strings.EqualFold(v, "delete") && len(arr) > k+2 {
			tableName = arr[k+2]
			break
		}

		// Selects can be "awkward".
		if strings.EqualFold(v, "from") && len(arr) > k+1 {
			if strings.Contains(arr[k+1], "(") || strings.Contains(arr[k+1], "select") {
				awkwardScenario = true
			} else {
				tableName = arr[k+1]
			}

			break
		}
	}

	if awkwardScenario {
		if database == "" {
			return strings.ToUpper(operation)
		}
		return fmt.Sprintf("%s %s", strings.ToUpper(operation), database)
	}

	// If we couldn't infer the table name, no database name was provided and
	// the operation couldn't be inferred (this could be the case of a random
	// string), then just flag it in the span name.
	if database == "" && tableName == "" {
		if operation == "unknown" {
			return "Unknown database operation"
		}

		return strings.ToUpper(operation)
	}

	if database == "" {
		return fmt.Sprintf("%s %s", strings.ToUpper(operation), tableName)
	}

	return fmt.Sprintf("%s %s.%s", strings.ToUpper(operation), database, tableName)
}
