package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prionis/dns-server/internal/database"
)

func (s Server) loggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				s.logger.Info(fmt.Sprintf("%s %s from %s",
					r.Method, r.URL.String(), r.RemoteAddr))
			}()
			next.ServeHTTP(w, r)
		})
		return http.HandlerFunc(fn)
	}
}

func (s Server) timeoutMiddleware(t time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), t)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
					s.logger.Error("timeout connection with " + r.RemoteAddr)
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (s Server) authorizationMiddleware(allowed []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(database.User)
			if !ok {
				s.logger.Error("can't get user from context for authorization")
				http.Error(w, "Internal error, try later", http.StatusInternalServerError)
				return
			}

			if !slices.Contains(allowed, user.Role) {
				s.logger.Error("user " + user.Login + " don't have rights")
				http.Error(w, "Not enough rights for this", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s Server) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			s.logger.Error("getting jwt token cookie from request: " + err.Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			s.logger.Error("JWT_SECRET environment variable is not set")
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			s.logger.Error("can't parse JWT token: " + err.Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			s.logger.Error("can't retrive claims from token: " + err.Error())
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		login := claims["login"].(string)
		user, err := s.db.GetUser(context.Background(), login)
		if err != nil {
			s.logger.Error("can't retrive user from database: " + err.Error())
			http.Error(w, "Internal error, try later", http.StatusInternalServerError)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "user", user))
		next.ServeHTTP(w, r)
	})
}
