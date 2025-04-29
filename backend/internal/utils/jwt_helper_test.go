package utils_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/constants"
	"github.com/ctrlb-hq/ctrlb-control-plane/backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessTokenAndValidate(t *testing.T) {
	email := "test@example.com"
	token, err := utils.GenerateAccessToken(email)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	// Validate the token
	subject, err := utils.ValidateJWT(token, "access")
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}
	if subject != email {
		t.Errorf("expected subject %v, got %v", email, subject)
	}
}

func TestGenerateRefreshTokenAndValidate(t *testing.T) {
	email := "refresh@example.com"
	token, err := utils.GenerateRefreshToken(email)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	// Validate the refresh token
	subject, err := utils.ValidateJWT(token, "refresh")
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}
	if subject != email {
		t.Errorf("expected subject %v, got %v", email, subject)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	email := "refresh-success@example.com"
	refreshToken, err := utils.GenerateRefreshToken(email)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	newAccessToken, err := utils.RefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("failed to refresh token: %v", err)
	}

	// Validate the new access token
	subject, err := utils.ValidateJWT(newAccessToken, "access")
	if err != nil {
		t.Fatalf("failed to validate new access token: %v", err)
	}
	if subject != email {
		t.Errorf("expected subject %v, got %v", email, subject)
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	email := "invalid-signature@example.com"
	token, err := utils.GenerateAccessToken(email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Corrupt the token
	invalidToken := token + "tampered"
	_, err = utils.ValidateJWT(invalidToken, "access")
	if err == nil {
		t.Fatal("expected validation to fail for tampered token, but it succeeded")
	}
}

func TestValidateJWT_WrongTokenType(t *testing.T) {
	email := "wrong-type@example.com"
	refreshToken, err := utils.GenerateRefreshToken(email)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	_, err = utils.ValidateJWT(refreshToken, "access")
	if err == nil || !strings.Contains(err.Error(), "invalid token type") {
		t.Fatalf("expected invalid token type error, got: %v", err)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	// Manually create an expired token
	expiredTime := time.Now().Add(-time.Minute) // 1 min ago
	claims := jwt.MapClaims{
		"sub":       "expired@example.com",
		"exp":       expiredTime.Unix(),
		"iat":       time.Now().Unix(),
		"token_use": "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(constants.JWT_SECRET))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	_, err = utils.ValidateJWT(tokenString, "access")
	if err == nil {
		t.Fatal("expected validation to fail for expired token, but it succeeded")
	}
}
