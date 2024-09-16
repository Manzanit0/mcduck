package servers_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/jmoiron/sqlx"
	usersv1 "github.com/manzanit0/mcduck/api/users.v1"
	"github.com/manzanit0/mcduck/cmd/dots/servers"
	"github.com/manzanit0/mcduck/internal/pgtest"
	"github.com/manzanit0/mcduck/internal/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()

	dbContainer, err := pgtest.NewDBContainer(ctx)
	require.NoError(t, err)

	connectionString, err := dbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sqlx.Open("pgx", connectionString)
	require.NoError(t, err)

	userEmail := "foo@email.com"
	telegramID := int64(1234)
	_, err = users.Create(ctx, db, users.User{Email: userEmail, Password: "foo", TelegramChatID: &telegramID})
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)

	err = dbContainer.Snapshot(ctx, postgres.WithSnapshotName("get_user"))
	require.NoError(t, err)

	t.Cleanup(func() {
		err = dbContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	t.Run("get existing user", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("get_user"))
			require.NoError(t, err)
		})

		s := servers.NewUsersServer(db)
		res, err := s.GetUser(ctx, &connect.Request[usersv1.GetUserRequest]{
			Msg: &usersv1.GetUserRequest{
				TelegramChatId: telegramID,
			},
		})

		require.NoError(t, err)
		assert.Equal(t, userEmail, res.Msg.User.Email)
		assert.EqualValues(t, telegramID, res.Msg.User.TelegramChatId)
	})

	t.Run("get non-existing user", func(t *testing.T) {
		db, err := sqlx.Open("pgx", connectionString)
		require.NoError(t, err)

		t.Cleanup(func() {
			err = db.Close()
			require.NoError(t, err)

			err = dbContainer.Restore(ctx, postgres.WithSnapshotName("get_user"))
			require.NoError(t, err)
		})

		s := servers.NewUsersServer(db)
		res, err := s.GetUser(ctx, &connect.Request[usersv1.GetUserRequest]{
			Msg: &usersv1.GetUserRequest{
				TelegramChatId: 88888,
			},
		})

		require.ErrorContains(t, err, "not_found: user with telegram ID 88888 doesn't exist")
		assert.Nil(t, res)
	})
}
