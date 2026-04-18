package httpauth

import (
	"log"
	"net/http"
	"strings"

	jwtmanager "architecture/backend/internal/shared/auth/jwt"
	"architecture/backend/internal/shared/auth/userctx"
	"architecture/backend/internal/shared/transport/httpjson"
)

func JWTAuth(manager *jwtmanager.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				log.Printf("auth: missing Authorization header path=%s", r.URL.Path)
				httpjson.Write(w, http.StatusUnauthorized, httpjson.ErrorResponse{Error: "Authorization header is required"})
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				log.Printf("auth: invalid scheme path=%s", r.URL.Path)
				httpjson.Write(w, http.StatusUnauthorized, httpjson.ErrorResponse{Error: "invalid authorization scheme"})
				return
			}

			userID, err := manager.Parse(strings.TrimSpace(parts[1]))
			if err != nil {
				log.Printf("auth: invalid token path=%s error=%v", r.URL.Path, err)
				httpjson.Write(w, http.StatusUnauthorized, httpjson.ErrorResponse{Error: "invalid token"})
				return
			}

			log.Printf("auth: success path=%s user_id=%s", r.URL.Path, userID)

			next.ServeHTTP(w, r.WithContext(userctx.WithUserID(r.Context(), userID)))
		})
	}
}
