package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"
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
		if utils.IsUniqueViolation(err) {
			return utils.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to register user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (a *AuthRepository) Login(email string) (*User, error) {
	var user User

	query := `SELECT email, name, password FROM user WHERE email = ?`
	err := a.db.QueryRow(query, email).Scan(&user.Email, &user.Name, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

func (a *AuthRepository) UserExists(email string) bool {
	var count int
	query := `SELECT COUNT(*) FROM user WHERE email = ?`

	err := a.db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
