package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		bearerToken := req.Header.Get("Authorization")
		if bearerToken == "" || strings.Contains(bearerToken, "Bearer") == false {
			err := errors.New("authenticate(): invalid authorization header")
			log.Ctx(req.Context()).Error().Err(err).Msg("")
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		token, err := firebaseAuth.VerifyIDToken(context.Background(), strings.TrimPrefix(bearerToken, "Bearer "))
		if err != nil {
			log.Ctx(req.Context()).Error().Err(errors.Wrap(err, "authenticate()")).Msg("invalid id token")
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), "userID", token.UID)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

type loggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (res *loggerResponseWriter) WriteHeader(statusCode int) {
	res.statusCode = statusCode
	res.ResponseWriter.WriteHeader(statusCode)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		lres := loggerResponseWriter{res, http.StatusOK}
		start := time.Now()

		log := log.With().Str("req_id", uuid.New().String()).Logger()
		ctx := log.WithContext(req.Context())

		hostname, err := os.Hostname()
		if err != nil {
			hostname = req.Host
			log.Error().Err(errors.Wrap(err, "logger()")).Msg("failed to get hostname for logger")
		}

		log.Info().
			Str("method", req.Method).
			Str("path", req.URL.Path).
			Str("query", req.URL.RawQuery).
			Str("client_ip", req.RemoteAddr).
			Str("user_agent", req.UserAgent()).
			Str("hostname", hostname).
			Msg("request received")

		defer func() {
			log.Info().
				Int("status_code", lres.statusCode).
				Dur("res_time", time.Since(start)).
				Msg("request completed")
		}()

		next.ServeHTTP(&lres, req.WithContext(ctx))
	})
}
