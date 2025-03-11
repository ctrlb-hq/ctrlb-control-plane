package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
	"github.com/mattn/go-sqlite3"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (a *AuthRepository) RegisterUser(user User) error {
	// Use transaction to handle user creation
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Attempt to insert user
	_, err = tx.Exec("INSERT INTO user (email, name, password, role) VALUES (?, ?, ?,?)", user.Email, user.Name, user.Password, user.Role)
	if err != nil {
		if isUniqueViolation(err) {
			log.Println("User already exists:", user)
			return utils.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to register user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("User registered:", user)
	return nil
}

func (a *AuthRepository) Login(email string) (*User, error) {
	var user User

	// Prepare and execute the SQL statement
	stmt, err := a.db.Prepare("SELECT email, name, password FROM user WHERE email = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	// Execute the query
	err = stmt.QueryRow(email).Scan(&user.Email, &user.Name, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	return &user, nil
}

func isUniqueViolation(err error) bool {
	if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
		return true
	}
	return false
}
