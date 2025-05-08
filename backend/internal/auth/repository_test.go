package auth

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite driver for testing
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE user (
			email TEXT PRIMARY KEY,
			name TEXT,
			password TEXT,
			role TEXT
		);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestRegisterUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	user := User{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
		Role:     "admin",
	}

	err := repo.RegisterUser(user)
	assert.NoError(t, err, "user should be registered without error")

	// Registering same user again should fail
	err = repo.RegisterUser(user)
	assert.Error(t, err, "duplicate registration should throw error")
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	// First, register a user
	user := User{
		Email:    "login@example.com",
		Name:     "Login User",
		Password: "securepass",
		Role:     "user",
	}
	err := repo.RegisterUser(user)
	assert.NoError(t, err)

	// Now try to login
	loggedInUser, err := repo.Login("login@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.Email, loggedInUser.Email)
	assert.Equal(t, user.Name, loggedInUser.Name)
	assert.Equal(t, user.Password, loggedInUser.Password)
	assert.Equal(t, user.Role, loggedInUser.Role)

	// Try to login with non-existent user
	_, err = repo.Login("nonexistent@example.com")
	assert.Error(t, err)
}

func TestUserExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	user := User{
		Email:    "exist@example.com",
		Name:     "Exist User",
		Password: "password",
		Role:     "viewer",
	}

	exists := repo.UserExists(user.Email)
	assert.False(t, exists, "user should not exist before registration")

	err := repo.RegisterUser(user)
	assert.NoError(t, err)

	exists = repo.UserExists(user.Email)
	assert.True(t, exists, "user should exist after registration")
}
