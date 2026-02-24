package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/url-shortener-go/internal/metrics"
)

// MetricsMiddleware собирает метрики по запросам
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter, чтобы перехватить статус
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.status)
		endpoint := r.URL.Path // можно улучшить, группируя по шаблонам

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, endpoint, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
