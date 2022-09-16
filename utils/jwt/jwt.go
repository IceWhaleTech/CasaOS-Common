package jwt

import (
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
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

// get AccessToken
func GetAccessToken(username, pwd string, id int) string {
	token, err := GenerateToken(username, pwd, id, "casaos", 3*time.Hour*time.Duration(1))
	if err == nil {
		return token
	}
	logger.Error("Get Token Fail", zap.Any("error", err))
	return ""
}

func GetRefreshToken(username, pwd string, id int) string {
	token, err := GenerateToken(username, pwd, id, "refresh", 7*24*time.Hour*time.Duration(1))
	if err == nil {
		return token
	}
	logger.Error("Get Token Fail", zap.Any("error", err))
	return ""
}

func Validate(token string) (*Claims, int) {
	if token == "" {
		return nil, common_err.INVALID_PARAMS
	}

	claims, err := ParseToken(token, false)

	if err != nil {
		return nil, common_err.ERROR_AUTH_TOKEN
	} else if !claims.VerifyExpiresAt(time.Now(), true) || !claims.VerifyIssuer("casaos", true) {
		return nil, common_err.ERROR_AUTH_TOKEN
	}

	return claims, common_err.SUCCESS
}
