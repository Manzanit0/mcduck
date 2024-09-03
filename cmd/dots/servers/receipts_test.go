package servers_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	receiptsv1 "github.com/manzanit0/mcduck/api/receipts.v1"
	"github.com/manzanit0/mcduck/cmd/dots/servers"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/internal/users"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"

	"connectrpc.com/connect"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestCreateReceipt(t *testing.T) {
	ctx := context.Background()

	dbContainer, err := NewDBContainer(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = dbContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	// We need to create the database connection AFTER the snapshot is taken.
	// Postgres errors but testcontainers-go silences that error. Pending
	// looking into the actual error.
	err = dbContainer.Snapshot(ctx)
	require.NoError(t, err)

	connectionString, err := dbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sqlx.Open("pgx", connectionString)
	require.NoError(t, err)

	t.Run("receipt is successfully created", func(t *testing.T) {
		t.Cleanup(func() {
			err = dbContainer.Restore(ctx)
			require.NoError(t, err)
		})

		parserClient := client.NewMockParserClient(t)
		tgramClient := tgram.NewMockClient(t)
		s := servers.NewReceiptsServer(db, parserClient, tgramClient)

		userEmail := "user@email.com"
		receiptBytes := []byte("foo")
		parserClient.EXPECT().
			ParseReceipt(mock.Anything, userEmail, receiptBytes).
			Return(&client.ParseReceiptResponse{
				Amount:       5.5,
				Currency:     "EUR",
				Description:  "some description",
				Vendor:       "some vendor",
				PurchaseDate: "02/01/2006",
			}, nil).
			Once()

		_, err = users.Create(ctx, db, users.User{Email: userEmail, Password: "foo"})
		require.NoError(t, err)

		ctx = auth.WithInfo(ctx, userEmail)
		res, err := s.CreateReceipt(ctx, &connect.Request[receiptsv1.CreateReceiptRequest]{
			Msg: &receiptsv1.CreateReceiptRequest{
				ReceiptFiles: [][]byte{receiptBytes},
			},
		})
		require.NoError(t, err)

		ids := res.Msg.GetReceiptIds()
		require.Len(t, ids, 1)

		r := receipt.NewRepository(db)
		receipt, err := r.GetReceipt(ctx, ids[0])
		require.NoError(t, err)
		assert.Equal(t, receipt.Vendor, "some vendor")
		assert.False(t, receipt.PendingReview)
		assert.Equal(t, receipt.Date.Format("02/01/2006"), "02/01/2006")

		e := expense.NewRepository(db)

		expenses, err := e.ListExpensesForReceipt(ctx, ids[0])
		require.NoError(t, err)
		require.Len(t, expenses, 1)
		assert.Equal(t, expenses[0].Amount, 5.5)
		assert.Equal(t, expenses[0].Category, "Receipt Upload")
		assert.Equal(t, expenses[0].Subcategory, "")
		assert.Equal(t, expenses[0].Description, "some description")
		assert.Equal(t, expenses[0].Date.Format("02/01/2006"), receipt.Date.Format("02/01/2006"))
	})
}

func NewDBContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	migrations, err := GetMigrationsFiles()
	if err != nil {
		return nil, fmt.Errorf("get migration files: %w", err)
	}

	container, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithInitScripts(migrations...),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
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
