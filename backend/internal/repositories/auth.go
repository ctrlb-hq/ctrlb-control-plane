package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/models"
)

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (a *AuthRepository) RegisterUser(user models.User) error {
	_, err := a.db.Exec("INSERT INTO user(email,name, password) VALUES(? , ?, ?)", user.Email, user.Name, user.Password)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: user.Email" {
			log.Println("User already exists:", user)
			return models.ErrUserAlreadyExists
		}
		return errors.New("failed to register user")
	}
	log.Println("User register:", user)
	return nil
}

func (a *AuthRepository) Login(email string, password string) (*models.User, error) {
	var user models.User

	// Prepare the SQL statement
	stmt, err := a.db.Prepare("SELECT email, name, password FROM user WHERE email = ?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %v", err)
	}
	defer stmt.Close()

	// Execute the query
	err = stmt.QueryRow(email).Scan(&user.Email, &user.Name, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	if user.Password != password {
		return nil, errors.New("invalid password")
	}

	return &user, nil
}
