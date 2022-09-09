package jwt

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/IceWhaleTech/CasaOS-Common/model"
	"github.com/IceWhaleTech/CasaOS-Common/utils/common_err"
	"github.com/IceWhaleTech/CasaOS-Common/utils/logger"
	"go.uber.org/zap"
)

func ExceptLocalhost2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]
		if clientIP == "::1" || clientIP == "127.0.0.1" {
			logger.Info("Bypassing JWT validation for request from localhost.", zap.Any("client_ip", r.RemoteAddr))
			h.ServeHTTP(w, r)
			return
		}

		JWT2(h).ServeHTTP(w, r)
	})
}

func JWT2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) == 0 {
			token = r.URL.Query().Get("token")
		}

		claims, code := validate(token)

		if code != common_err.SUCCESS {

			// serialize the response
			res, err := json.Marshal(model.Result{Success: code, Message: common_err.GetMsg(code)})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write(res); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}
		r.Header.Add("user_id", strconv.Itoa(claims.ID))
		h.ServeHTTP(w, r)
	})
}
