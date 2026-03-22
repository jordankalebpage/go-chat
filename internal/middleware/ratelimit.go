package middleware

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

type IPRateLimiter struct {
	visitors sync.Map
	limit    rate.Limit
	burst    int
	ttl      time.Duration
	requests atomic.Uint64
}

func NewIPRateLimiter(limit rate.Limit, burst int, ttl time.Duration) *IPRateLimiter {
	return &IPRateLimiter{
		limit: limit,
		burst: burst,
		ttl:   ttl,
	}
}

func (l *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r.RemoteAddr)
		limiter := l.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		l.pruneIfNeeded()
		next.ServeHTTP(w, r)
	})
}

func (l *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	loaded, ok := l.visitors.Load(ip)
	if ok {
		entry := loaded.(*visitor)
		entry.lastSeen.Store(time.Now().UnixNano())
		return entry.limiter
	}

	entry := &visitor{
		limiter: rate.NewLimiter(l.limit, l.burst),
	}

	entry.lastSeen.Store(time.Now().UnixNano())

	actual, _ := l.visitors.LoadOrStore(ip, entry)

	return actual.(*visitor).limiter
}

func (l *IPRateLimiter) pruneIfNeeded() {
	requestNumber := l.requests.Add(1)
	if requestNumber%128 != 0 {
		return
	}

	cutoff := time.Now().Add(-l.ttl).UnixNano()

	l.visitors.Range(func(key any, value any) bool {
		entry := value.(*visitor)

		if entry.lastSeen.Load() < cutoff {
			l.visitors.Delete(key)
		}

		return true
	})
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return host
}
