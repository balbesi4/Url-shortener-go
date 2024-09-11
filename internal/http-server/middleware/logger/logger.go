package logger

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/logger"),
		)
		log.Info("logger middleware initialized")

		function := func(writer http.ResponseWriter, request *http.Request) {
			entry := log.With(
				slog.String("method", request.Method),
				slog.String("url", request.URL.Path),
				slog.String("remote", request.RemoteAddr),
				slog.String("user-agent", request.UserAgent()),
				slog.String("request_id", middleware.GetReqID(request.Context())),
			)
			wrapWriter := middleware.NewWrapResponseWriter(writer, request.ProtoMajor)

			timeStart := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.String("status", strconv.Itoa(wrapWriter.Status())),
					slog.String("duration", time.Since(timeStart).String()),
					slog.String("bytes_written", strconv.Itoa(wrapWriter.BytesWritten())),
				)
			}()

			next.ServeHTTP(wrapWriter, request)
		}

		return http.HandlerFunc(function)
	}
}
