package auth

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type Claims struct {
	UserID string `json:"ui"`
	jwt.RegisteredClaims
}

func getTokenSecretVar(key string) []byte {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return []byte(os.Getenv(key))
}

func GenerateAccessToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	accessTokenSecret := getTokenSecretVar("ACCESS_TOKEN_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessTokenSecret)
}

func GenerateRefreshToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}

	refreshTokenSecret := getTokenSecretVar("REFRESH_TOKEN_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshTokenSecret)
}

func ValidateToken(tokenString string, isRefresh bool) (*Claims, error) {
	accessTokenSecret := getTokenSecretVar("ACCESS_TOKEN_SECRET")
	refreshTokenSecret := getTokenSecretVar("REFRESH_TOKEN_SECRET")
	secret := accessTokenSecret
	if isRefresh {
		secret = refreshTokenSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func RefreshTokens(refreshTokenString string) (string, string, error) {
	claims, err := ValidateToken(refreshTokenString, true)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err := GenerateAccessToken(claims.UserID)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := GenerateRefreshToken(claims.UserID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
