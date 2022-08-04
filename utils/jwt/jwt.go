package jwt

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Username string `json:"username"`
	Password string `json:"password"`
	ID       int    `json:"id"`
	jwt.RegisteredClaims
}

func GenerateToken(username, password string, id int, issuer string, t time.Duration) (string, error) {
	var secret []byte // TODO: need to use some global secret accessible by all CasaOS services

	claims := Claims{
		username,
		password,
		id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(secret)
	return token, err
}

func ParseToken(token string, valid bool) (*Claims, error) {
	var secret []byte // TODO: need to use some global secret accessible by all CasaOS services

	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok {
			if valid && tokenClaims.Valid {
				return claims, nil
			} else if !valid {
				return claims, nil
			}
		}
	}
	return nil, err
}
