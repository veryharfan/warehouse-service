package pkg

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	UID int64  `json:"uid"`
	SID *int64 `json:"sid"`
}

func ParseJwtToken(tokenString string, secretKey string) (TokenClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return TokenClaims{}, err
	}

	// Validate the token and extract claims
	var tokenClaims TokenClaims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["uid"].(float64); ok {
			tokenClaims.UID = int64(uid)
		}
		if shopID, ok := claims["sid"].(float64); ok {
			tokenClaims.SID = new(int64)
			*tokenClaims.SID = int64(shopID)
		}
		return tokenClaims, nil
	}

	return TokenClaims{}, fmt.Errorf("invalid token claims")
}

func GetTokenFromHeaders(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("missing token")
	}

	// Split the header to get the token
	token := header[len("Bearer "):]
	if token == "" {
		return "", fmt.Errorf("invalid token")
	}

	return token, nil
}
