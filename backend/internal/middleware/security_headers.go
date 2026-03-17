package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds essential security headers to all responses
// Implements OWASP security best practices for HTTP headers
// env parameter should be the application environment (e.g., "production", "development")
func SecurityHeaders(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Frame-Options: Prevents clickjacking attacks
		// DENY: Page cannot be displayed in a frame/iframe
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevents MIME type sniffing
		// nosniff: Browser must respect declared content-type
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Enables XSS filter in older browsers
		// 1; mode=block: Enable filter and block page if attack detected
		// Note: Modern browsers use CSP instead, but this helps legacy browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security (HSTS): Enforces HTTPS
		// max-age=31536000: Browser should use HTTPS for 1 year
		// includeSubDomains: Apply to all subdomains
		// preload: Allow inclusion in browser HSTS preload lists
		// WARNING: Only enable in production with valid HTTPS certificate
		if env == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Content-Security-Policy (CSP): Prevents XSS and data injection
		// This is a restrictive policy suitable for API-only backend
		csp := "default-src 'none'; " + // Block all by default
			"frame-ancestors 'none'; " + // No framing allowed (redundant with X-Frame-Options but better)
			"base-uri 'self'; " + // Restrict base tag to same origin
			"form-action 'self'" // Forms can only submit to same origin
		c.Header("Content-Security-Policy", csp)

		// Referrer-Policy: Controls referrer information
		// no-referrer: Don't send referrer information at all
		c.Header("Referrer-Policy", "no-referrer")

		// Permissions-Policy (formerly Feature-Policy): Control browser features
		// Disable all features since this is an API backend
		permissions := "accelerometer=(), " +
			"camera=(), " +
			"geolocation=(), " +
			"gyroscope=(), " +
			"magnetometer=(), " +
			"microphone=(), " +
			"payment=(), " +
			"usb=()"
		c.Header("Permissions-Policy", permissions)

		// X-Permitted-Cross-Domain-Policies: Restrict cross-domain data access
		// none: No cross-domain policy files allowed
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Cache-Control: Prevent sensitive data caching
		// This is set globally here but can be overridden per-route
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Header("Pragma", "no-cache") // HTTP/1.0 compatibility
		c.Header("Expires", "0")       // Proxies

		// Server header: Remove or obfuscate server information
		// Don't advertise Go/Gin version to potential attackers
		c.Header("Server", "Ascend-API")

		c.Next()
	}
}

// SecureHeaders is an alias for SecurityHeaders for convenience
// Can be used interchangeably: router.Use(middleware.SecureHeaders(env))
var SecureHeaders = SecurityHeaders
