package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/gin-gonic/gin"
)

func ExceptLocalhost() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ClientIP() == "::1" || c.ClientIP() == "127.0.0.1" {
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

		claims, code := Validate(token)

		if code != common_err.SUCCESS {
			c.JSON(http.StatusUnauthorized, model.Result{Success: code, Message: common_err.GetMsg(code)})
			c.Abort()
			return
		}
		c.Request.Header.Add("user_id", strconv.Itoa(claims.ID))
		c.Next()
	}
}

func GenerateSecret() (string, error) {
	randomBytes := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(randomBytes), nil
}
