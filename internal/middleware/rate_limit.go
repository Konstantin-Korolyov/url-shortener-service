package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters = make(map[string]*ipLimiter)
	mu       sync.Mutex
	once     sync.Once // для однократного запуска очистки
)

const (
	rateLimit       = 10 // запросов в секунду
	burst           = 20 // максимальный всплеск
	cleanupInterval = 5 * time.Minute
	maxLimiters     = 10000 // ограничение размера карты
)

// RateLimit ограничивает количество запросов с одного IP
func RateLimit(next http.Handler) http.Handler {
	once.Do(func() {
		go func() {
			for {
				time.Sleep(cleanupInterval)
				mu.Lock()
				count := 0
				for ip, lim := range limiters {
					if time.Since(lim.lastSeen) > 10*time.Minute {
						delete(limiters, ip)
					} else {
						count++
						// При достижении лимита прекращаем перебор
						if count >= maxLimiters {
							break
						}
					}
				}
				mu.Unlock()
			}
		}()
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		mu.Lock()
		lim, ok := limiters[ip]
		if !ok {
			lim = &ipLimiter{
				limiter:  rate.NewLimiter(rate.Limit(rateLimit), burst),
				lastSeen: time.Now(),
			}
			limiters[ip] = lim
		}
		lim.lastSeen = time.Now()
		mu.Unlock()

		if !lim.limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
