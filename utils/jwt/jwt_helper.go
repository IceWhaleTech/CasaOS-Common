package jwt

import (
	"net/http"
	"strconv"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ExceptLocalhost() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ClientIP() == "::1" || c.ClientIP() == "127.0.0.1" {
			logger.Info("Bypassing JWT validation for request from localhost.", zap.Any("client_ip", c.ClientIP()))
			c.Next()
			return
		}

		JWT()(c)
	}
}

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) == 0 {
			token = c.Query("token")
		}

		claims, code := validate(token)

		if code != common_err.SUCCESS {
			c.JSON(http.StatusUnauthorized, model.Result{Success: code, Message: common_err.GetMsg(code)})
			c.Abort()
			return
		}
		c.Request.Header.Add("user_id", strconv.Itoa(claims.ID))
		c.Next()
	}
}
