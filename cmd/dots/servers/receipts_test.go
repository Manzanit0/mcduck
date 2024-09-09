package servers_test

import (
	"context"
	"testing"
	"time"

	receiptsv1 "github.com/manzanit0/mcduck/api/receipts.v1"
	"github.com/manzanit0/mcduck/cmd/dots/servers"
	"github.com/manzanit0/mcduck/internal/client"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/pgtest"
	"github.com/manzanit0/mcduck/internal/receipt"
	"github.com/manzanit0/mcduck/internal/users"
	"github.com/manzanit0/mcduck/pkg/auth"
	"github.com/manzanit0/mcduck/pkg/tgram"
	"google.golang.org/protobuf/types/known/timestamppb"

	"connectrpc.com/connect"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestCreateReceipt(t *testing.T) {
	ctx := context.Background()

	dbContainer, err := pgtest.NewDBContainer(ctx)
	require.NoError(t, err)

	connectionString, err := dbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	err = dbContainer.Snapshot(ctx, postgres.WithSnapshotName("create_receipt"))
	require.NoError(t, err)

	t.Cleanup(func() {
		err = dbContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	t.Run("when context doesn't have email, system panics", func(t *testing.T) {
		ctx = auth.WithInfo(ctx, "") // no email
		s := servers.NewReceiptsServer(nil, nil, nil)

		require.PanicsWithValue(t, "empty user email", func() {
			_, _ = s.CreateReceipt(ctx, &connect.Request[receiptsv1.CreateReceiptRequest]{
				Msg: &receiptsv1.CreateReceiptRequest{
					ReceiptFiles: [][]byte{[]byte("")},
				},
			})
		})
	})

	t.Run("receipt is successfully created", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
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
		assert.True(t, receipt.PendingReview)
		assert.Equal(t, receipt.Date.Format("02/01/2006"), "02/01/2006")

		e := expense.NewRepository(db)

		expenses, err := e.ListExpensesForReceipt(ctx, ids[0])
		require.NoError(t, err)
		require.Len(t, expenses, 1)
		assert.Equal(t, expenses[0].Amount, float32(5.5))
		assert.Equal(t, expenses[0].Category, "Receipt Upload")
		assert.Equal(t, expenses[0].Subcategory, "")
		assert.Equal(t, expenses[0].Description, "some description")
		assert.Equal(t, expenses[0].Date.Format("02/01/2006"), receipt.Date.Format("02/01/2006"))
	})

	t.Run("invalid dates are transformed to 'now'", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
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
				PurchaseDate: "some-gibberish", // invalid date
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
		assert.True(t, receipt.PendingReview)
		assert.Equal(t, receipt.Date.Format("02/01/2006"), time.Now().Format("02/01/2006"))

		e := expense.NewRepository(db)

		expenses, err := e.ListExpensesForReceipt(ctx, ids[0])
		require.NoError(t, err)
		require.Len(t, expenses, 1)
		assert.Equal(t, expenses[0].Amount, float32(5.5))
		assert.Equal(t, expenses[0].Category, "Receipt Upload")
		assert.Equal(t, expenses[0].Subcategory, "")
		assert.Equal(t, expenses[0].Description, "some description")
		assert.Equal(t, expenses[0].Date.Format("02/01/2006"), time.Now().Format("02/01/2006"))
	})

	t.Run("empty images are rejected", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
			require.NoError(t, err)
		})

		parserClient := client.NewMockParserClient(t)
		tgramClient := tgram.NewMockClient(t)
		s := servers.NewReceiptsServer(db, parserClient, tgramClient)

		userEmail := "user@email.com"
		receiptBytes := []byte("") // empty image
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
		_, err = s.CreateReceipt(ctx, &connect.Request[receiptsv1.CreateReceiptRequest]{
			Msg: &receiptsv1.CreateReceiptRequest{
				ReceiptFiles: [][]byte{receiptBytes},
			},
		})
		require.ErrorContains(t, err, "internal: create receipt: empty receipt")
	})

	t.Run("when a receipt fails to be created, an error is returned", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
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

		// Let's close the connection to force a DB error.
		err = db.Close()
		require.NoError(t, err)

		ctx = auth.WithInfo(ctx, userEmail)
		_, err = s.CreateReceipt(ctx, &connect.Request[receiptsv1.CreateReceiptRequest]{
			Msg: &receiptsv1.CreateReceiptRequest{
				ReceiptFiles: [][]byte{receiptBytes},
			},
		})
		require.ErrorContains(t, err, "internal: create receipt: begin transaction: sql: database is close")
	})
}

func TestUpdateReceipt(t *testing.T) {
	ctx := context.Background()

	dbContainer, err := pgtest.NewDBContainer(ctx)
	require.NoError(t, err)

	connectionString, err := dbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sqlx.Open("pgx", connectionString)
	require.NoError(t, err)

	userEmail := "foo@email.com"
	_, err = users.Create(ctx, db, users.User{Email: userEmail, Password: "foo"})
	require.NoError(t, err)

	repo := receipt.NewRepository(db)
	existingReceipt, err := repo.CreateReceipt(ctx, receipt.CreateReceiptRequest{
		Amount:      5,
		Description: "description",
		Vendor:      "vendor",
		Image:       []byte("foo"),
		Date:        time.Now(),
		Email:       userEmail,
	})
	require.NoError(t, err)

	// Note: Create receipt only returns the ID.
	existingReceipt, err = repo.GetReceipt(ctx, uint64(existingReceipt.ID))
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)

	err = dbContainer.Snapshot(ctx, postgres.WithSnapshotName("create_receipt"))
	require.NoError(t, err)

	t.Cleanup(func() {
		err = dbContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	t.Run("everything is updated", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
			require.NoError(t, err)
		})

		updateStr := "updated"
		updateBool := true
		updateDate := timestamppb.New(time.Date(1993, 2, 24, 0, 0, 0, 0, time.UTC))
		s := servers.NewReceiptsServer(db, nil, nil)
		_, err = s.UpdateReceipt(ctx, &connect.Request[receiptsv1.UpdateReceiptRequest]{
			Msg: &receiptsv1.UpdateReceiptRequest{
				Id:            uint64(existingReceipt.ID),
				Vendor:        &updateStr,
				PendingReview: &updateBool,
				Date:          updateDate,
			},
		})
		require.NoError(t, err)

		repo = receipt.NewRepository(db)
		updatedReceipt, err := repo.GetReceipt(ctx, uint64(existingReceipt.ID))
		assert.NoError(t, err)
		assert.Equal(t, updateStr, updatedReceipt.Vendor)
		assert.Equal(t, updateBool, updatedReceipt.PendingReview)
		assert.Equal(t, "24/02/1993", updatedReceipt.Date.Format("02/01/2006"))
		assert.Equal(t, existingReceipt.UserEmail, updatedReceipt.UserEmail)
	})

	t.Run("only vendor is updated", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
			require.NoError(t, err)
		})

		updateValue := "updated"
		s := servers.NewReceiptsServer(db, nil, nil)
		_, err = s.UpdateReceipt(ctx, &connect.Request[receiptsv1.UpdateReceiptRequest]{
			Msg: &receiptsv1.UpdateReceiptRequest{
				Id:     uint64(existingReceipt.ID),
				Vendor: &updateValue,
			},
		})
		require.NoError(t, err)

		repo = receipt.NewRepository(db)
		updatedReceipt, err := repo.GetReceipt(ctx, uint64(existingReceipt.ID))
		assert.NoError(t, err)
		assert.Equal(t, updateValue, updatedReceipt.Vendor)
		assert.Equal(t, existingReceipt.PendingReview, updatedReceipt.PendingReview)
		assert.Equal(t, existingReceipt.UserEmail, updatedReceipt.UserEmail)
		assert.Equal(t, existingReceipt.Date, updatedReceipt.Date)
	})

	t.Run("only pending review is updated", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
			require.NoError(t, err)
		})

		updateValue := true
		s := servers.NewReceiptsServer(db, nil, nil)
		_, err = s.UpdateReceipt(ctx, &connect.Request[receiptsv1.UpdateReceiptRequest]{
			Msg: &receiptsv1.UpdateReceiptRequest{
				Id:            uint64(existingReceipt.ID),
				PendingReview: &updateValue,
			},
		})
		require.NoError(t, err)

		repo = receipt.NewRepository(db)
		updatedReceipt, err := repo.GetReceipt(ctx, uint64(existingReceipt.ID))
		assert.NoError(t, err)
		assert.Equal(t, updateValue, updatedReceipt.PendingReview)
		assert.Equal(t, existingReceipt.Vendor, updatedReceipt.Vendor)
		assert.Equal(t, existingReceipt.UserEmail, updatedReceipt.UserEmail)
		assert.Equal(t, existingReceipt.Date, updatedReceipt.Date)
	})

	t.Run("only date is updated", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
			require.NoError(t, err)
		})

		updateValue := timestamppb.New(time.Date(1993, 2, 24, 0, 0, 0, 0, time.UTC))
		s := servers.NewReceiptsServer(db, nil, nil)
		_, err = s.UpdateReceipt(ctx, &connect.Request[receiptsv1.UpdateReceiptRequest]{
			Msg: &receiptsv1.UpdateReceiptRequest{
				Id:   uint64(existingReceipt.ID),
				Date: updateValue,
			},
		})
		require.NoError(t, err)

		repo = receipt.NewRepository(db)
		updatedReceipt, err := repo.GetReceipt(ctx, uint64(existingReceipt.ID))
		assert.NoError(t, err)
		assert.Equal(t, "24/02/1993", updatedReceipt.Date.Format("02/01/2006"))
		assert.Equal(t, existingReceipt.Vendor, updatedReceipt.Vendor)
		assert.Equal(t, existingReceipt.UserEmail, updatedReceipt.UserEmail)
		assert.Equal(t, existingReceipt.PendingReview, updatedReceipt.PendingReview)
	})

	t.Run("when invalid id is provided, error is returned", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_receipt"))
			require.NoError(t, err)
		})

		s := servers.NewReceiptsServer(db, nil, nil)
		_, err = s.UpdateReceipt(ctx, &connect.Request[receiptsv1.UpdateReceiptRequest]{
			Msg: &receiptsv1.UpdateReceiptRequest{
				Id: 123123,
			},
		})

		assert.ErrorContains(t, err, "invalid_argument: receipt with id 123123 doesn't exist")
	})
}
