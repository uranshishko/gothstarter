package middleware

import (
	"net/http"

	"github.com/uranshishko/gothstarter/auth"
	"github.com/uranshishko/gothstarter/common"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hc := common.NewHandlerContext(w, r)

		if _, err := auth.Client.CompleteAuth(hc); err != nil {
			hc.Response().Header().Set("HX-Redirect", "/login")
			hc.Redirect(302, "/login")
			return
		}

		next.ServeHTTP(w, r)
	})
}
