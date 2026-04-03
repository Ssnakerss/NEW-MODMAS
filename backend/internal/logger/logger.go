package logger

import (
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Sl struct {
	Log *slog.Logger
}

// Setup logger with specific log level basing on environment env variable
func Setup(env string, output io.Writer) *Sl {
	var l Sl
	switch env {
	case envLocal:
		l.Log = slog.New(
			slog.NewTextHandler(output, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		l.Log = slog.New(
			slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		l.Log = slog.New(
			slog.NewJSONHandler(output, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return &l
}

func (l *Sl) Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

// Middleware для оборачивания хэндлера и логирования событий
// Для обработки запросов
func (l *Sl) WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.URL
		method := r.Method
		//
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		// SLog.Infof("request header: %v", r.Header)
		//
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		l.Log.Info("request details:",
			"uri", uri,
			"method", method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}

// Теперь займемся ответами
type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter // оигинальный вритер
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}
