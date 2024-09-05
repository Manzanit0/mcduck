package pgtest

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName                    = "mcduck_db"
	dbUser                    = "postgres"
	dbPassword                = "mcduck_test_db_password"
	dbPort                    = "5432"
	migrationsDirRelativePath = "../../../migrations/"
)

func NewDBContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	migrations, err := GetMigrationsFiles()
	if err != nil {
		return nil, fmt.Errorf("get migration files: %w", err)
	}

	container, err := postgres.Run(ctx,
		"docker.io/postgres:15.8-alpine3.20",
		postgres.WithInitScripts(migrations...),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("run postgres testcontainer: %w", err)
	}

	return container, nil
}

func GetMigrationsFiles() ([]string, error) {
	var migrationsFiles []string
	files, err := os.ReadDir(migrationsDirRelativePath)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationsFiles = append(migrationsFiles, migrationsDirRelativePath+file.Name())
		}
	}

	return migrationsFiles, nil
}
