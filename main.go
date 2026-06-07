package main

import (
    "log"
    "net/http"
    "sync"
    "time"

    "golang.org/x/crypto/bcrypt"
)

// ---------- a) Хэширование паролей (уже есть) ----------
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// ---------- b) Защита от кликджекинга и другие заголовки ----------
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Frame-Options", "DENY")            // защита от кликджекинга
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}

// ---------- c) Самостоятельная мера: Rate Limiting ----------
type rateLimiter struct {
    mu     sync.Mutex
    visits map[string][]time.Time
    limit  int
    window time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
    return &rateLimiter{
        visits: make(map[string][]time.Time),
        limit:  limit,
        window: window,
    }
}

func (rl *rateLimiter) allow(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    now := time.Now()
    var recent []time.Time
    for _, t := range rl.visits[ip] {
        if now.Sub(t) < rl.window {
            recent = append(recent, t)
        }
    }
    if len(recent) >= rl.limit {
        return false
    }
    recent = append(recent, now)
    rl.visits[ip] = recent
    return true
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        if !rl.allow(ip) {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// ---------- d) HTTPS ----------
func main() {
    // Демонстрация хэширования (необязательно, но показывает работу)
    password := "examplePassword"
    hash, _ := hashPassword(password)
    log.Printf("Hash of '%s': %s", password, hash)
    log.Printf("Match: %v", checkPasswordHash(password, hash))

    // Роутер
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Secure DevSecOps App"))
    })
    mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })

    // Применяем middleware: rate limiting -> security headers
    handler := securityHeaders(mux)
    rl := newRateLimiter(5, time.Second) // 5 запросов в секунду на IP
    handler = rl.middleware(handler)

    // Запуск HTTP (порт 8080) – для проверки
    go func() {
        log.Println("HTTP server listening on :8080")
        if err := http.ListenAndServe(":8080", handler); err != nil {
            log.Fatal(err)
        }
    }()

    // Запуск HTTPS (порт 8443) – требуется cert.pem и key.pem
    log.Println("HTTPS server listening on :8443")
    if err := http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", handler); err != nil {
        log.Fatal(err)
    }
}
