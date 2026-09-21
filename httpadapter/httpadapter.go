// Package httpadapter serves i18n-kit's language detection over net/http.
//
// It lives in its own package for the same reason fiberadapter does: importing
// the root package should not drag a web server into a binary that never
// serves one. Translation lookup is just as useful in a CLI printing localized
// help or error messages, and net/http costs such a binary well over a hundred
// packages it has no use for.
//
// Everything here is a translation layer. What a language is, how a request is
// read, how a cookie value is resolved, and what a missing key formats to are
// all decided in the root package and read from there, so net/http and Fiber
// can never drift apart. It is the net/http half of the contract
// i18n.RequestSource describes.
package httpadapter

import (
	"net/http"

	i18n "github.com/soulteary/i18n-kit/v4"
)

// RequestSourceOf adapts an *http.Request to i18n.RequestSource.
//
// Exported so that code holding a detector of its own can reuse the adapter
// without going through Detect.
func RequestSourceOf(r *http.Request) i18n.RequestSource { return stdSource{r: r} }

type stdSource struct{ r *http.Request }

func (s stdSource) Query(name string) string {
	if s.r.URL == nil {
		return ""
	}
	return s.r.URL.Query().Get(name)
}

func (s stdSource) Cookie(name string) string {
	cookie, err := s.r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s stdSource) Header(name string) string { return s.r.Header.Get(name) }

// Detect detects the language of a request using the default detector.
func Detect(r *http.Request) i18n.Language {
	return DetectWith(i18n.DefaultDetector, r)
}

// DetectWith detects the language of a request using the given detector.
func DetectWith(d *i18n.Detector, r *http.Request) i18n.Language {
	return d.Detect(RequestSourceOf(r))
}

// SetLanguage returns a copy of r carrying lang in its context.
func SetLanguage(r *http.Request, lang i18n.Language) *http.Request {
	return r.WithContext(i18n.ContextWithLanguage(r.Context(), lang))
}

// Language extracts the language from the request context.
// Returns i18n.DefaultLanguage if not found.
func Language(r *http.Request) i18n.Language {
	if r == nil {
		return i18n.DefaultLanguage
	}
	return i18n.LanguageFromContext(r.Context())
}

// T returns the translated string using the language from the request context.
func T(r *http.Request, key string) string {
	if r == nil {
		return i18n.TFromContext(nil, key)
	}
	return i18n.TFromContext(r.Context(), key)
}

// Tf returns a formatted translated string using the language from the request
// context.
//
// If the key has no translation, the key is returned as-is and the arguments
// are NOT applied -- the root package decides that, not this one.
func Tf(r *http.Request, key string, args ...interface{}) string {
	if r == nil {
		return i18n.TfFromContext(nil, key, args...)
	}
	return i18n.TfFromContext(r.Context(), key, args...)
}

// SameSite maps a resolved cookie mode to net/http's SameSite value.
//
// It is a function here rather than a method on i18n.CookieSameSiteMode
// because a method would have to live in the root package, and its return type
// is exactly the net/http dependency this package exists to keep out of it.
// The second result is false for i18n.SameSiteDisabled, meaning write no
// attribute at all.
func SameSite(m i18n.CookieSameSiteMode) (http.SameSite, bool) {
	switch m {
	case i18n.SameSiteLax:
		return http.SameSiteLaxMode, true
	case i18n.SameSiteStrict:
		return http.SameSiteStrictMode, true
	case i18n.SameSiteNone:
		return http.SameSiteNoneMode, true
	default:
		return 0, false
	}
}

// Config configures the net/http middleware.
//
// It embeds i18n.MiddlewareConfig -- everything framework-independent -- and
// adds only what net/http needs, exactly as fiberadapter.Config does for Fiber.
type Config struct {
	i18n.MiddlewareConfig

	// Next skips this middleware when it returns true.
	Next func(r *http.Request) bool
}

// Middleware creates a net/http middleware for language detection.
// It is the net/http counterpart of fiberadapter.Middleware.
func Middleware(config ...Config) func(http.Handler) http.Handler {
	var next func(r *http.Request) bool
	var base []i18n.MiddlewareConfig
	if len(config) > 0 {
		next = config[0].Next
		base = []i18n.MiddlewareConfig{config[0].MiddlewareConfig}
	}
	cfg := i18n.ResolveMiddlewareConfig(base...)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if next != nil && next(r) {
				h.ServeHTTP(w, r)
				return
			}

			lang := DetectWith(cfg.Detector, r)

			ctx := i18n.ContextWithLanguage(r.Context(), lang)
			if cfg.Bundle != nil {
				ctx = i18n.ContextWithBundle(ctx, cfg.Bundle)
			}
			r = r.WithContext(ctx)

			if cfg.SetCookie {
				mode := i18n.ResolveCookieSameSite(cfg.CookieSameSite)

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
				if sameSite, ok := SameSite(mode); ok {
					cookie.SameSite = sameSite
				}

				http.SetCookie(w, cookie)
			}

			h.ServeHTTP(w, r)
		})
	}
}

// MiddlewareFunc is Middleware for wrapping an http.HandlerFunc directly.
func MiddlewareFunc(config ...Config) func(http.HandlerFunc) http.HandlerFunc {
	middleware := Middleware(config...)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return middleware(next).ServeHTTP
	}
}

// SimpleMiddleware creates a middleware that only detects language.
// It is a convenience for the common use case.
func SimpleMiddleware() func(http.Handler) http.Handler {
	return Middleware(Config{MiddlewareConfig: i18n.DefaultMiddlewareConfig()})
}
