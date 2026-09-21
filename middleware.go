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

	// DisableCookieHTTPOnly clears the HttpOnly flag on the cookie.
	//
	// Negative so that the zero value is the documented default. As a plain
	// CookieHTTPOnly bool it could not be told apart from "not set", and the
	// merge worked around that by guessing: it took the caller's value only
	// once CookieName or CookieSameSite was also set, so naming the cookie and
	// nothing else silently cleared HttpOnly. There is nothing to guess now.
	//
	// Default: false, i.e. the cookie is HttpOnly.
	DisableCookieHTTPOnly bool

	// CookieSameSite sets the SameSite attribute: "Lax", "Strict", "None" or
	// "disabled" to omit the attribute. Matching is case-insensitive and an
	// unrecognised value means "Lax"; see ResolveCookieSameSite, which every
	// middleware uses so the frameworks cannot read it differently.
	//
	// "None" forces Secure on, because browsers reject the combination
	// otherwise and the cookie is never stored.
	//
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
// storage, so that every adapter and every reader of one agrees on where to
// look.
//
// They are NOT interchangeable with this package's net/http context keys. Those
// spell the same two strings but have an unexported type, so a value stored
// under LocalsLanguageKey does not read back through LanguageFromContext, and
// TFromContext on such a framework answers in the default language without
// reporting anything. Read the language through the adapter's own accessor --
// fiberadapter.Language, say -- or call ContextWithLanguage yourself first if
// you want the context helpers to see it.
const (
	LocalsLanguageKey = "i18n-language"
	LocalsBundleKey   = "i18n-bundle"
)

// mergeConfig merges user config with defaults.
// User-provided non-zero values override defaults. Every bool is taken from the
// caller as-is: each one is named so that its zero value is the default, which
// is what lets the merge be this plain.
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
	// DisableCookieHTTPOnly is a bool - take from user
	result.DisableCookieHTTPOnly = user.DisableCookieHTTPOnly
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
				mode := ResolveCookieSameSite(cfg.CookieSameSite)

				cookie := &http.Cookie{
					Name:     cfg.CookieName,
					Value:    string(lang),
					MaxAge:   cfg.CookieMaxAge,
					Path:     cfg.CookiePath,
					Secure:   cfg.CookieSecure || mode.RequiresSecure(),
					HttpOnly: !cfg.DisableCookieHTTPOnly,
				}
				// A zero SameSite writes no attribute, which is what
				// SameSiteDisabled asks for.
				if sameSite, ok := mode.HTTPSameSite(); ok {
					cookie.SameSite = sameSite
				}

				http.SetCookie(w, cookie)
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
