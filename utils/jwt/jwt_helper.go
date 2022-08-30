package jwt

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func JWT(bypassLocalhost bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bypassLocalhost {
			if c.ClientIP() == "::1" || c.ClientIP() == "127.0.0.1" {
				logger.Info("Bypassing JWT validation because request comes from localhost", zap.Any("ClientIP", c.ClientIP()))
				c.Next()
				return
			}
		}

		var code int
		code = common_err.SUCCESS
		token := c.GetHeader("Authorization")
		if len(token) == 0 {
			token = c.Query("token")
		}
		if token == "" {
			code = common_err.INVALID_PARAMS
		}

		claims, err := ParseToken(token, false)

		//_, err := ParseToken(token)
		if err != nil {
			code = common_err.ERROR_AUTH_TOKEN
		} else if (c.Request.URL.Path == "/v1/file" || c.Request.URL.Path == "/v1/image" || c.Request.URL.Path == "/v1/file/upload" || c.Request.URL.Path == "/v1/batch") && claims.VerifyIssuer("casaos", true) {
			// Special treatment
		} else if !claims.VerifyExpiresAt(time.Now(), true) || !claims.VerifyIssuer("casaos", true) {
			code = common_err.ERROR_AUTH_TOKEN
		}
		if code != common_err.SUCCESS {
			c.JSON(http.StatusUnauthorized, model.Result{Success: code, Message: common_err.GetMsg(code)})
			c.Abort()
			return
		}
		c.Request.Header.Add("user_id", strconv.Itoa(claims.ID))
		c.Next()
	}
}

// get AccessToken
func GetAccessToken(username, pwd string, id int) string {
	token, err := GenerateToken(username, pwd, id, "casaos", 3*time.Hour*time.Duration(1))
	if err == nil {
		return token
	}
	logger.Error(fmt.Sprintf("Get Token Fail: %V", err))
	return ""
}

func GetRefreshToken(username, pwd string, id int) string {
	token, err := GenerateToken(username, pwd, id, "refresh", 7*24*time.Hour*time.Duration(1))
	if err == nil {
		return token
	}
	logger.Error(fmt.Sprintf("Get Token Fail: %V", err))
	return ""
}
