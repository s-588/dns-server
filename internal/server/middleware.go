package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prionis/dns-server/internal/database"
)

func (s Server) loggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				slog.Info("Request", "method", r.Method, "path", r.URL.String(), "addr",
					r.RemoteAddr)
			}()
			next.ServeHTTP(w, r)
		})
		return fn
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
					slog.Error("connection timeout", "addr", r.RemoteAddr)
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
				slog.Error("can't get user from context for authorization")
				http.Error(w, "Internal error, try later", http.StatusInternalServerError)
				return
			}

			if !slices.Contains(allowed, user.Role) {
				slog.Error("user " + user.Login + " don't have rights")
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
			slog.Error("can't retrieve jwt token from request", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			slog.Error("JWT_SECRET environment variable is not set")
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			slog.Error("can't parse JWT token", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			slog.Error("can't retrieve claims from token", "error", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		login, ok := claims["login"].(string)
		if !ok {
			slog.Error("can't retrieve login from claims")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user, err := s.db.GetUser(context.Background(), login)
		if err != nil {
			slog.Error("can't retrieve user from database", "error", err.Error())
			http.Error(w, "Internal error, try later", http.StatusInternalServerError)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), contextKeyUser, user))
		next.ServeHTTP(w, r)
	})
}

func prometheusMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			status := strconv.Itoa(ww.Status())
			path := r.URL.Path

			m.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
		})
	}
}
