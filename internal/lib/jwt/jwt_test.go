package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"

	"github.com/CVSaby/sso-service/internal/domain/models"
)

func TestNewToken(t *testing.T) {
	user := models.User{
		ID:    123,
		Email: "test@example.com",
	}
	duration := time.Hour
	secret := "mysecret"

	// Тест успешного создания токена
	tokenString, err := NewToken(user, duration, secret)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if tokenString == "" {
		t.Fatal("Expected non-empty token string")
	}

	// Проверка декодирования токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		t.Fatalf("Expected valid token, got error: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["uid"] != float64(user.ID) {
			t.Errorf("Expected uid %d, got %v", user.ID, claims["uid"])
		}
		if claims["email"] != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, claims["email"])
		}
		if claims["exp"] == nil {
			t.Error("Expected exp claim to exist")
		}
	} else {
		t.Error("Token is invalid")
	}
}

func TestNewTokenError(t *testing.T) {
	user := models.User{
		ID:    123,
		Email: "test@example.com",
	}
	duration := time.Hour
	secret := "" // пустой секрет должен вызвать ошибку

	_, err := NewToken(user, duration, secret)

	if err == nil {
		t.Fatal("Expected error due to empty secret, got nil")
	}
}
