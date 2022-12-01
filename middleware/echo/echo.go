package middleware

import (
	"strconv"

	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	"github.com/labstack/echo/v4"
	echo_middleware "github.com/labstack/echo/v4/middleware"
)

var CommonCORSConfiguration = echo_middleware.CORSConfig{
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{echo.POST, echo.GET, echo.OPTIONS, echo.PUT, echo.DELETE},
	AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentLength, echo.HeaderXCSRFToken, echo.HeaderContentType, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders, echo.HeaderAccessControlAllowMethods, echo.HeaderConnection, echo.HeaderOrigin, echo.HeaderXRequestedWith},
	ExposeHeaders:    []string{echo.HeaderContentLength, echo.HeaderAccessControlAllowOrigin, echo.HeaderAccessControlAllowHeaders},
	MaxAge:           172800,
	AllowCredentials: true,
}

var CommonJWTConfiguration = echo_middleware.JWTConfig{
	Skipper: func(c echo.Context) bool {
		if c.RealIP() == "::1" || c.RealIP() == "127.0.0.1" {
			return true
		}

		if c.Request().Method == echo.GET && c.Request().Header.Get(echo.HeaderUpgrade) == "websocket" {
			return true
		}

		return false
	},
	ParseTokenFunc: func(token string, c echo.Context) (interface{}, error) {
		claims, code := jwt.Validate(token)
		if code != common_err.SUCCESS {
			return nil, echo.ErrUnauthorized
		}

		c.Request().Header.Set("user_id", strconv.Itoa(claims.ID))

		return claims, nil
	},
	TokenLookupFuncs: []echo_middleware.ValuesExtractor{
		func(c echo.Context) ([]string, error) {
			return []string{
				c.Request().Header.Get(echo.HeaderAuthorization),
				c.Request().URL.Query()["token"][0],
			}, nil
		},
	},
}
