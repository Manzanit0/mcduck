package servers_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	expensesv1 "github.com/manzanit0/mcduck/api/expenses.v1"
	"github.com/manzanit0/mcduck/cmd/dots/servers"
	"github.com/manzanit0/mcduck/internal/expense"
	"github.com/manzanit0/mcduck/internal/pgtest"
	"github.com/manzanit0/mcduck/internal/users"
	"google.golang.org/protobuf/types/known/timestamppb"

	"connectrpc.com/connect"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestUpdateExpense(t *testing.T) {
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

	err = db.Close()
	require.NoError(t, err)

	err = dbContainer.Snapshot(ctx, postgres.WithSnapshotName("create_expense"))
	require.NoError(t, err)

	t.Cleanup(func() {
		err = dbContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	t.Run("only amount is changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateAmount := float32(1500)
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:     uint64(expenseID),
				Amount: &updateAmount,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1500, res.Msg.Expense.Amount)
		assert.Empty(t, res.Msg.Expense.Category)
		assert.Empty(t, res.Msg.Expense.Subcategory)
		assert.Empty(t, res.Msg.Expense.Description)
		assert.Nil(t, res.Msg.Expense.ReceiptId)
	})

	t.Run("only category is changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateCategory := "Travel"
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:       uint64(expenseID),
				Category: &updateCategory,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1200, res.Msg.Expense.Amount)
		assert.EqualValues(t, "Travel", res.Msg.Expense.Category)
		assert.Empty(t, res.Msg.Expense.Subcategory)
		assert.Empty(t, res.Msg.Expense.Description)
		assert.Nil(t, res.Msg.Expense.ReceiptId)
	})

	t.Run("only subcategory is changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateSubcategory := "Flight"
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:          uint64(expenseID),
				Subcategory: &updateSubcategory,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1200, res.Msg.Expense.Amount)
		assert.Empty(t, res.Msg.Expense.Category)
		assert.EqualValues(t, "Flight", res.Msg.Expense.Subcategory)
		assert.Empty(t, res.Msg.Expense.Description)
		assert.Nil(t, res.Msg.Expense.ReceiptId)
	})

	t.Run("only description is changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateDescription := "Business trip to NYC"
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:          uint64(expenseID),
				Description: &updateDescription,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1200, res.Msg.Expense.Amount)
		assert.Empty(t, res.Msg.Expense.Category)
		assert.Empty(t, res.Msg.Expense.Subcategory)
		assert.EqualValues(t, "Business trip to NYC", res.Msg.Expense.Description)
		assert.Nil(t, res.Msg.Expense.ReceiptId)
	})

	t.Run("only receipt ID is changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateReceiptID := uint64(123456)
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:        uint64(expenseID),
				ReceiptId: &updateReceiptID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1200, res.Msg.Expense.Amount)
		assert.Empty(t, res.Msg.Expense.Category)
		assert.Empty(t, res.Msg.Expense.Subcategory)
		assert.Empty(t, res.Msg.Expense.Description)
		assert.EqualValues(t, 123456, *res.Msg.Expense.ReceiptId)
	})

	t.Run("all fields are changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateDate := timestamppb.New(time.Date(1993, 2, 24, 0, 0, 0, 0, time.UTC))
		updateAmount := float32(1500)
		updateCategory := "Travel"
		updateSubcategory := "Flight"
		updateDescription := "Business trip to NYC"
		updateReceiptID := uint64(123456)
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:          uint64(expenseID),
				Date:        updateDate,
				Amount:      &updateAmount,
				Category:    &updateCategory,
				Subcategory: &updateSubcategory,
				Description: &updateDescription,
				ReceiptId:   &updateReceiptID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1500, res.Msg.Expense.Amount)
		assert.EqualValues(t, "Travel", res.Msg.Expense.Category)
		assert.EqualValues(t, "Flight", res.Msg.Expense.Subcategory)
		assert.EqualValues(t, "Business trip to NYC", res.Msg.Expense.Description)
		assert.EqualValues(t, 123456, *res.Msg.Expense.ReceiptId)
		assert.EqualValues(t, "24/02/1993", res.Msg.Expense.Date.AsTime().Format("02/01/2006"))
	})

	t.Run("when expense doesn't exist, server returns error", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		s := servers.NewExpensesServer(db)

		_, err = s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id: 9000,
			},
		})

		assert.ErrorContains(t, err, "invalid_argument: expense with id 9000 doesn't exist")
	})

	t.Run("only date is changed", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("create_expense"))
			require.NoError(t, err)
		})

		repo := expense.NewRepository(db)
		expenseID, err := repo.CreateExpense(ctx, expense.CreateExpenseRequest{
			UserEmail: userEmail,
			Date:      time.Now(),
			Amount:    12,
			ReceiptID: nil,
		})
		require.NoError(t, err)

		s := servers.NewExpensesServer(db)

		updateDate := timestamppb.New(time.Date(1993, 2, 24, 0, 0, 0, 0, time.UTC))
		res, err := s.UpdateExpense(ctx, &connect.Request[expensesv1.UpdateExpenseRequest]{
			Msg: &expensesv1.UpdateExpenseRequest{
				Id:   uint64(expenseID),
				Date: updateDate,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Msg.Expense)
		assert.EqualValues(t, expenseID, res.Msg.Expense.Id)
		assert.EqualValues(t, 1200, res.Msg.Expense.Amount)
		assert.Empty(t, res.Msg.Expense.Category)
		assert.Empty(t, res.Msg.Expense.Subcategory)
		assert.Empty(t, res.Msg.Expense.Description)
		assert.Nil(t, res.Msg.Expense.ReceiptId)
		assert.EqualValues(t, "24/02/1993", res.Msg.Expense.Date.AsTime().Format("02/01/2006"))
	})
}
