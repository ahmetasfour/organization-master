package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersMiddleware adds security headers to all responses.
// Implements common security best practices similar to helmet.js.
func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// X-Frame-Options: Prevents clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevents MIME-type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Legacy XSS protection (modern browsers)
		c.Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Controls referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// X-Permitted-Cross-Domain-Policies: Restricts Adobe Flash/PDF policies
		c.Set("X-Permitted-Cross-Domain-Policies", "none")

		// X-Download-Options: Prevents IE from executing downloads in site's context
		c.Set("X-Download-Options", "noopen")

		// Strict-Transport-Security: Enforces HTTPS (only effective over HTTPS)
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content-Security-Policy: Basic policy for API responses
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// Permissions-Policy: Restricts browser features
		c.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

		return c.Next()
	}
}

// ─── Rate Limiting ─────────────────────────────────────────────────────────────

// RateLimiterConfig configures the rate limiter.
type RateLimiterConfig struct {
	// MaxRequests is the maximum number of requests allowed within the window.
	MaxRequests int
	// Window is the time window for rate limiting.
	Window time.Duration
}

// rateLimitEntry tracks request counts per IP.
type rateLimitEntry struct {
	count     int
	expiresAt time.Time
}

// rateLimiter stores rate limit state.
type rateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*rateLimitEntry
	config  RateLimiterConfig
}

// newRateLimiter creates a new rate limiter with the given configuration.
func newRateLimiter(config RateLimiterConfig) *rateLimiter {
	rl := &rateLimiter{
		entries: make(map[string]*rateLimitEntry),
		config:  config,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// allow checks if the request is allowed and increments the counter.
func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[ip]

	if !exists || now.After(entry.expiresAt) {
		// New entry or expired — reset
		rl.entries[ip] = &rateLimitEntry{
			count:     1,
			expiresAt: now.Add(rl.config.Window),
		}
		return true
	}

	// Within window — check limit
	if entry.count >= rl.config.MaxRequests {
		return false
	}

	entry.count++
	return true
}

// cleanup removes expired entries periodically.
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.After(entry.expiresAt) {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware with the given configuration.
func RateLimitMiddleware(config RateLimiterConfig) fiber.Handler {
	limiter := newRateLimiter(config)

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if !limiter.allow(ip) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please try again later.",
				},
			})
		}
		return c.Next()
	}
}

// PublicTokenRateLimiter returns rate limiting middleware for public token endpoints.
// Allows 10 requests per minute per IP.
func PublicTokenRateLimiter() fiber.Handler {
	return RateLimitMiddleware(RateLimiterConfig{
		MaxRequests: 10,
		Window:      time.Minute,
	})
}

// LoginRateLimiter returns rate limiting middleware for login endpoint.
// Allows 5 requests per minute per IP.
func LoginRateLimiter() fiber.Handler {
	return RateLimitMiddleware(RateLimiterConfig{
		MaxRequests: 5,
		Window:      time.Minute,
	})
}
