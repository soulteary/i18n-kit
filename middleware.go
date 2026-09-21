package i18n

import (
	"net/http"
)

// MiddlewareConfig configures the i18n middleware.
type MiddlewareConfig struct {
	// Detector is the language detector to use.
	// Default: DefaultDetector
	Detector *Detector

	// Bundle is the translation bundle to inject into context.
	// If nil, only language detection is performed.
	Bundle *Bundle

	// SetCookie enables setting a language cookie after detection.
	// Default: false
	SetCookie bool

	// CookieName is the name of the cookie to set.
	// Default: "lang"
	CookieName string

	// CookieMaxAge is the max age of the cookie in seconds.
	// Default: 86400 * 365 (1 year)
	CookieMaxAge int

	// CookiePath is the path for the cookie.
	// Default: "/"
	CookiePath string

	// CookieSecure sets the Secure flag on the cookie.
	// Default: false
	CookieSecure bool

	// CookieHTTPOnly sets the HttpOnly flag on the cookie.
	// Default: true
	CookieHTTPOnly bool

	// CookieSameSite sets the SameSite attribute.
	// Default: "Lax"
	CookieSameSite string

	// NextStd defines a function to skip middleware for net/http when true.
	// Default: nil
	NextStd func(r *http.Request) bool
}

// DefaultMiddlewareConfig returns the default middleware configuration.
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		Detector:       DefaultDetector,
		Bundle:         nil,
		SetCookie:      false,
		CookieName:     "lang",
		CookieMaxAge:   86400 * 365,
		CookiePath:     "/",
		CookieSecure:   false,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
		NextStd:        nil,
	}
}

// ResolveMiddlewareConfig turns a middleware's variadic config argument into
// the effective configuration, applying the same defaults and the same merge
// rules every middleware uses. Framework adapters call this instead of
// restating the rules -- see the fiberadapter subpackage.
func ResolveMiddlewareConfig(config ...MiddlewareConfig) MiddlewareConfig {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = mergeConfig(cfg, config[0])
	}
	return cfg
}

// LocalsLanguageKey and LocalsBundleKey are the keys under which a framework
// adapter stores the detected language and the bundle on its per-request
// storage. They match the net/http context keys, so a value set by one
// framework reads back the same way everywhere.
const (
	LocalsLanguageKey = "i18n-language"
	LocalsBundleKey   = "i18n-bundle"
)

// mergeConfig merges user config with defaults.
// User-provided non-zero values override defaults.
// For boolean fields like SetCookie and CookieSecure, we always take the user's value.
// For CookieHTTPOnly, we keep the default (true) unless user explicitly provides cookie config.
func mergeConfig(defaults, user MiddlewareConfig) MiddlewareConfig {
	result := defaults

	if user.Detector != nil {
		result.Detector = user.Detector
	}
	if user.Bundle != nil {
		result.Bundle = user.Bundle
	}
	// SetCookie is a bool that we always take from user
	result.SetCookie = user.SetCookie
	if user.CookieName != "" {
		result.CookieName = user.CookieName
	}
	if user.CookieMaxAge != 0 {
		result.CookieMaxAge = user.CookieMaxAge
	}
	if user.CookiePath != "" {
		result.CookiePath = user.CookiePath
	}
	// CookieSecure is a bool - take from user
	result.CookieSecure = user.CookieSecure
	// CookieHTTPOnly: We keep the default (true) unless user explicitly provides other cookie settings
	// This is a pragmatic choice since HttpOnly=true is almost always what you want for security
	// If user provides CookieName or CookieSameSite, we assume they want full control
	if user.CookieName != "" || user.CookieSameSite != "" {
		result.CookieHTTPOnly = user.CookieHTTPOnly
	}
	if user.CookieSameSite != "" {
		result.CookieSameSite = user.CookieSameSite
	}
	if user.NextStd != nil {
		result.NextStd = user.NextStd
	}

	return result
}

// StdMiddleware creates a net/http middleware for language detection.
func StdMiddleware(config ...MiddlewareConfig) func(http.Handler) http.Handler {
	cfg := ResolveMiddlewareConfig(config...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip middleware if NextStd returns true
			if cfg.NextStd != nil && cfg.NextStd(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Detect language
			lang := cfg.Detector.DetectFromRequest(r)

			// Store in context
			ctx := ContextWithLanguage(r.Context(), lang)

			// Store bundle if provided
			if cfg.Bundle != nil {
				ctx = ContextWithBundle(ctx, cfg.Bundle)
			}

			// Update request with new context
			r = r.WithContext(ctx)

			// Set cookie if configured
			if cfg.SetCookie {
				sameSite := http.SameSiteLaxMode
				switch cfg.CookieSameSite {
				case "Strict":
					sameSite = http.SameSiteStrictMode
				case "None":
					sameSite = http.SameSiteNoneMode
				}

				http.SetCookie(w, &http.Cookie{
					Name:     cfg.CookieName,
					Value:    string(lang),
					MaxAge:   cfg.CookieMaxAge,
					Path:     cfg.CookiePath,
					Secure:   cfg.CookieSecure,
					HttpOnly: cfg.CookieHTTPOnly,
					SameSite: sameSite,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StdMiddlewareFunc creates a net/http middleware function for language detection.
// This is useful when you need to wrap a http.HandlerFunc directly.
func StdMiddlewareFunc(config ...MiddlewareConfig) func(http.HandlerFunc) http.HandlerFunc {
	middleware := StdMiddleware(config...)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return middleware(next).ServeHTTP
	}
}

// SimpleMiddleware creates a simple middleware that only detects language.
// This is a convenience function for the common use case.
func SimpleMiddleware() func(http.Handler) http.Handler {
	return StdMiddleware(DefaultMiddlewareConfig())
}
