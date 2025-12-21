package logger

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(data) // записываем данные в оригинальный ResponseWriter
	lrw.responseData.size += size

	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode) // передаём статус код в оригинальный ResponseWriter
	lrw.responseData.status = statusCode
}

func NewLoggerMiddleware(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var resp responseData

			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &resp,
			}

			next.ServeHTTP(lw, r)

			log.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Dur("duration", time.Since(start)).
				Msg("Got request")

			log.Info().
				Int("status", lw.responseData.status).
				Int("size", lw.responseData.size).
				Msg("Send response")
		})
	}
}
