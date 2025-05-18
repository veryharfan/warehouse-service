package pkg

import (
	"time"

	"github.com/golang-jwt/jwt"
)

func CreateJwtToken(userId int64, secretKey string, expired int64) (string, error) {
	// Create a new JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": userId,
		"exp": jwt.TimeFunc().Add(time.Second * time.Duration(expired)).Unix(), // Token expiration time
		"iat": jwt.TimeFunc().Unix(),                                           // Token issued at time
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
