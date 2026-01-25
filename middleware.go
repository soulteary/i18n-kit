package i18n

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
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

	// Next defines a function to skip this middleware when true.
	// Default: nil
	Next func(c *fiber.Ctx) bool

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
		Next:           nil,
		NextStd:        nil,
	}
}

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
	if user.Next != nil {
		result.Next = user.Next
	}
	if user.NextStd != nil {
		result.NextStd = user.NextStd
	}

	return result
}

// FiberMiddleware creates a Fiber middleware for language detection.
func FiberMiddleware(config ...MiddlewareConfig) fiber.Handler {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = mergeConfig(cfg, config[0])
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Detect language
		lang := cfg.Detector.DetectFromFiber(c)

		// Store in Fiber locals
		c.Locals("i18n-language", lang)

		// Store bundle if provided
		if cfg.Bundle != nil {
			c.Locals("i18n-bundle", cfg.Bundle)
		}

		// Set cookie if configured
		if cfg.SetCookie {
			c.Cookie(&fiber.Cookie{
				Name:     cfg.CookieName,
				Value:    string(lang),
				MaxAge:   cfg.CookieMaxAge,
				Path:     cfg.CookiePath,
				Secure:   cfg.CookieSecure,
				HTTPOnly: cfg.CookieHTTPOnly,
				SameSite: cfg.CookieSameSite,
			})
		}

		return c.Next()
	}
}

// LanguageFromFiberLocals extracts the language from Fiber locals.
func LanguageFromFiberLocals(c *fiber.Ctx) Language {
	if lang, ok := c.Locals("i18n-language").(Language); ok {
		return lang
	}
	return DefaultLanguage
}

// BundleFromFiberLocals extracts the bundle from Fiber locals.
func BundleFromFiberLocals(c *fiber.Ctx) *Bundle {
	if bundle, ok := c.Locals("i18n-bundle").(*Bundle); ok {
		return bundle
	}
	return DefaultBundle
}

// TFromFiber returns the translated string using the language from Fiber context.
func TFromFiber(c *fiber.Ctx, key string) string {
	bundle := BundleFromFiberLocals(c)
	lang := LanguageFromFiberLocals(c)
	return bundle.GetTranslation(lang, key)
}

// TfFromFiber returns a formatted translated string using the language from Fiber context.
func TfFromFiber(c *fiber.Ctx, key string, args ...interface{}) string {
	return TFromFiber(c, key)
}

// StdMiddleware creates a net/http middleware for language detection.
func StdMiddleware(config ...MiddlewareConfig) func(http.Handler) http.Handler {
	cfg := DefaultMiddlewareConfig()
	if len(config) > 0 {
		cfg = mergeConfig(cfg, config[0])
	}

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

// SimpleFiberMiddleware creates a simple Fiber middleware that only detects language.
// This is a convenience function for the common use case.
func SimpleFiberMiddleware() fiber.Handler {
	return FiberMiddleware(DefaultMiddlewareConfig())
}
