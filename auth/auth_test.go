package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWT_GenerateToken(t *testing.T) {
	secret := []byte("test-secret-key-32-chars!!")
	j := NewJWT(secret, time.Hour)

	claims := jwt.MapClaims{
		"sub":  "user123",
		"name": "John",
	}

	token, err := j.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if !parsed.Valid {
		t.Error("expected valid token")
	}
}

func TestJWT_GenerateToken_ContainsClaims(t *testing.T) {
	secret := []byte("test-secret-key-32-chars!!")
	j := NewJWT(secret, time.Hour)

	claims := jwt.MapClaims{
		"sub":  "user123",
		"name": "John",
	}

	token, err := j.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mapClaims := parsed.Claims.(jwt.MapClaims)
	if mapClaims["sub"] != "user123" {
		t.Errorf("expected sub=user123, got %v", mapClaims["sub"])
	}
	if mapClaims["name"] != "John" {
		t.Errorf("expected name=John, got %v", mapClaims["name"])
	}
}

func TestJWT_GenerateToken_Expiry(t *testing.T) {
	secret := []byte("test-secret-key-32-chars!!")
	j := NewJWT(secret, time.Hour)

	claims := jwt.MapClaims{}

	token, err := j.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mapClaims := parsed.Claims.(jwt.MapClaims)
	exp, ok := mapClaims["exp"].(float64)
	if !ok {
		t.Error("expected exp claim to be present")
	}
	if exp == 0 {
		t.Error("expected exp to be set")
	}
}
