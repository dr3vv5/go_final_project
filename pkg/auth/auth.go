package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	// 1. Берем секретный ключ ДЛЯ ПОДПИСИ ТОКЕНА из отдельной переменной
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		// Для локального теста можно оставить дефолт, но помни: в продакшене это ошибка!
		key = "change-me-to-random-string-in-production"
	}
	jwtSecret = []byte(key)
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func GenerateToken(password string) (string, error) {
	// 1. Считаем хеш пароля (ты это уже сделал)
	h := sha256.Sum256([]byte(password))
	payloadHash := hex.EncodeToString(h[:])

	// 2. Создаем полезную нагрузку (Claims)
	claims := &Claims{
		PasswordHash: payloadHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)), // Токен живет 8 часов
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 3. Создаем сам токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 4. Подписываем токен секретным ключом и получаем строку
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func ValidateToken(tokenString string, currentPassword string) (bool, error) {
	h := sha256.Sum256([]byte(currentPassword))
	currentHash := hex.EncodeToString(h[:])

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return false, err
	}

	if claims.PasswordHash != currentHash {
		return false, errors.New("token invalid due to password change")
	}

	return true, nil
}
