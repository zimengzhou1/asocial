package integration

import (
	"asocial/internal/db"
	"asocial/internal/domain"
	"asocial/internal/repository"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *db.DB {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	cfg := db.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "asocial",
		Password: "asocial_dev_password",
		DBName:   "asocial",
		SSLMode:  "disable",
	}

	database, err := db.New(cfg, logger)
	require.NoError(t, err, "Failed to connect to database")

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

func cleanupUsers(t *testing.T, database *db.DB) {
	_, err := database.Exec("DELETE FROM room_user_settings")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM rooms")
	require.NoError(t, err)
	_, err = database.Exec("DELETE FROM users")
	require.NoError(t, err)
}

func TestDatabaseConnection(t *testing.T) {
	database := setupTestDB(t)

	err := database.HealthCheck()
	assert.NoError(t, err, "Database health check should pass")

	err = database.Ping()
	assert.NoError(t, err, "Database ping should succeed")
}

func TestUserRepository_Create(t *testing.T) {
	database := setupTestDB(t)
	cleanupUsers(t, database)

	repo := repository.NewUserRepository(database.DB)
	ctx := context.Background()

	t.Run("create new user successfully", func(t *testing.T) {
		params := domain.CreateUserParams{
			Email:    "test@example.com",
			Username: "testuser",
		}

		user, err := repo.Create(ctx, params)
		require.NoError(t, err)

		assert.NotNil(t, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "testuser", user.Username)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
		assert.False(t, user.LastSeenAt.IsZero())
	})

	t.Run("duplicate email should fail", func(t *testing.T) {
		cleanupUsers(t, database)

		params1 := domain.CreateUserParams{
			Email:    "duplicate@example.com",
			Username: "user1",
		}
		_, err := repo.Create(ctx, params1)
		require.NoError(t, err)

		params2 := domain.CreateUserParams{
			Email:    "duplicate@example.com",
			Username: "user2",
		}
		_, err = repo.Create(ctx, params2)
		assert.Error(t, err, "Duplicate email should fail")
	})

	t.Run("duplicate username should fail", func(t *testing.T) {
		cleanupUsers(t, database)

		params1 := domain.CreateUserParams{
			Email:    "user1@example.com",
			Username: "duplicateuser",
		}
		_, err := repo.Create(ctx, params1)
		require.NoError(t, err)

		params2 := domain.CreateUserParams{
			Email:    "user2@example.com",
			Username: "duplicateuser",
		}
		_, err = repo.Create(ctx, params2)
		assert.Error(t, err, "Duplicate username should fail")
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	database := setupTestDB(t)
	cleanupUsers(t, database)

	repo := repository.NewUserRepository(database.DB)
	ctx := context.Background()

	params := domain.CreateUserParams{
		Email:    "getbyid@example.com",
		Username: "getbyiduser",
	}

	created, err := repo.Create(ctx, params)
	require.NoError(t, err)

	t.Run("get existing user by ID", func(t *testing.T) {
		user, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, created.Email, user.Email)
		assert.Equal(t, created.Username, user.Username)
	})

	t.Run("get non-existent user returns nil", func(t *testing.T) {
		fakeID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
		user, err := repo.GetByID(ctx, fakeID)
		require.NoError(t, err)
		assert.Nil(t, user, "Getting non-existent user should return nil")
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	database := setupTestDB(t)
	cleanupUsers(t, database)

	repo := repository.NewUserRepository(database.DB)
	ctx := context.Background()

	params := domain.CreateUserParams{
		Email:    "getbyemail@example.com",
		Username: "getemailuser",
	}

	created, err := repo.Create(ctx, params)
	require.NoError(t, err)

	t.Run("get existing user by email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "getbyemail@example.com")
		require.NoError(t, err)

		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, created.Email, user.Email)
	})

	t.Run("get non-existent email returns nil", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		require.NoError(t, err)
		assert.Nil(t, user, "Getting non-existent email should return nil")
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	database := setupTestDB(t)
	cleanupUsers(t, database)

	repo := repository.NewUserRepository(database.DB)
	ctx := context.Background()

	params := domain.CreateUserParams{
		Email:    "getbyusername@example.com",
		Username: "getusernameuser",
	}

	created, err := repo.Create(ctx, params)
	require.NoError(t, err)

	t.Run("get existing user by username", func(t *testing.T) {
		user, err := repo.GetByUsername(ctx, "getusernameuser")
		require.NoError(t, err)

		assert.Equal(t, created.ID, user.ID)
		assert.Equal(t, created.Username, user.Username)
	})

	t.Run("get non-existent username returns nil", func(t *testing.T) {
		user, err := repo.GetByUsername(ctx, "nonexistentuser")
		require.NoError(t, err)
		assert.Nil(t, user, "Getting non-existent username should return nil")
	})
}

func TestUserRepository_UpdateLastSeen(t *testing.T) {
	database := setupTestDB(t)
	cleanupUsers(t, database)

	repo := repository.NewUserRepository(database.DB)
	ctx := context.Background()

	params := domain.CreateUserParams{
		Email:    "lastseen@example.com",
		Username: "lastseenuser",
	}

	created, err := repo.Create(ctx, params)
	require.NoError(t, err)

	originalLastSeen := created.LastSeenAt
	time.Sleep(100 * time.Millisecond)

	err = repo.UpdateLastSeenAt(ctx, created.ID)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)

	assert.True(t, updated.LastSeenAt.After(originalLastSeen),
		"LastSeenAt should be updated to a later time")
}

func TestRoomRepository_Create(t *testing.T) {
	database := setupTestDB(t)
	cleanupUsers(t, database)

	roomRepo := repository.NewRoomRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)
	ctx := context.Background()

	// Create a user first
	userParams := domain.CreateUserParams{
		Email:    "roomowner@example.com",
		Username: "roomowner",
	}
	user, err := userRepo.Create(ctx, userParams)
	require.NoError(t, err)

	t.Run("create room with owner", func(t *testing.T) {
		params := domain.CreateRoomParams{
			Name:        "Test Room",
			Slug:        "test-room",
			Description: stringPtr("A test room"),
			OwnerID:     &user.ID,
			IsPublic:    true,
		}

		room, err := roomRepo.Create(ctx, params)
		require.NoError(t, err)

		assert.NotNil(t, room.ID)
		assert.Equal(t, "Test Room", room.Name)
		assert.Equal(t, "test-room", room.Slug)
		assert.NotNil(t, room.OwnerID)
		assert.Equal(t, user.ID, *room.OwnerID)
		assert.True(t, room.IsPublic)
	})

	t.Run("duplicate slug should fail", func(t *testing.T) {
		params1 := domain.CreateRoomParams{
			Name:     "Room 1",
			Slug:     "duplicate-slug",
			IsPublic: true,
		}
		_, err := roomRepo.Create(ctx, params1)
		require.NoError(t, err)

		params2 := domain.CreateRoomParams{
			Name:     "Room 2",
			Slug:     "duplicate-slug",
			IsPublic: true,
		}
		_, err = roomRepo.Create(ctx, params2)
		assert.Error(t, err, "Duplicate slug should fail")
	})
}

func stringPtr(s string) *string {
	return &s
}
