// src/services/jwt.go

package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func GenerateJWT(userID int, username string, secretKey []byte, expiry time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiry)

	claims := &JWTClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "fxm",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJWT(tokenString string, secretKey []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token错误")
}
