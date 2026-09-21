// Package fiberadapter wires i18n-kit into Fiber v3.
//
// It lives in its own package so that importing the root package does not drag
// Fiber -- and with it fasthttp -- into binaries that never use it. A service
// on net/http, Echo, Gin or chi pays nothing for Fiber support existing; only
// importing this package links it in.
//
// Everything here is a translation layer. Detection order, config merging and
// the translation-formatting rule all live in the root package and are read
// from there, so Fiber and net/http cannot drift apart.
package fiberadapter

import (
	"github.com/gofiber/fiber/v3"

	i18n "github.com/soulteary/i18n-kit/v3"
)

// Source adapts a fiber.Ctx to i18n.RequestSource. Detection itself stays in
// the root package; this only says where the three values come from.
type Source struct{ C fiber.Ctx }

// Query returns a URL query parameter, or "" when absent.
func (s Source) Query(name string) string { return s.C.Query(name) }

// Cookie returns a request cookie's value, or "" when absent.
func (s Source) Cookie(name string) string { return s.C.Cookies(name) }

// Header returns a request header, or "" when absent.
func (s Source) Header(name string) string { return s.C.Get(name) }

// Detect detects the language of a Fiber request using the package-level
// default detector. It is the Fiber counterpart of i18n.DetectFromRequest.
func Detect(c fiber.Ctx) i18n.Language {
	return i18n.DefaultDetector.Detect(Source{C: c})
}

// DetectWith detects using a specific detector.
func DetectWith(d *i18n.Detector, c fiber.Ctx) i18n.Language {
	return d.Detect(Source{C: c})
}

// Config is i18n.MiddlewareConfig plus the Fiber-typed skip predicate.
//
// Next cannot live on i18n.MiddlewareConfig: a func(fiber.Ctx) bool field is
// exactly what pulled Fiber into the root package in the first place.
type Config struct {
	i18n.MiddlewareConfig

	// Next skips this middleware when it returns true.
	Next func(c fiber.Ctx) bool
}

// Middleware creates a Fiber middleware for language detection.
// It is the Fiber counterpart of i18n.StdMiddleware.
func Middleware(config ...Config) fiber.Handler {
	var next func(c fiber.Ctx) bool
	var base []i18n.MiddlewareConfig
	if len(config) > 0 {
		next = config[0].Next
		base = append(base, config[0].MiddlewareConfig)
	}
	cfg := i18n.ResolveMiddlewareConfig(base...)

	return func(c fiber.Ctx) error {
		if next != nil && next(c) {
			return c.Next()
		}

		lang := cfg.Detector.Detect(Source{C: c})
		c.Locals(i18n.LocalsLanguageKey, lang)

		if cfg.Bundle != nil {
			c.Locals(i18n.LocalsBundleKey, cfg.Bundle)
		}

		if cfg.SetCookie {
			// Normalised in the root package rather than here: Fiber takes the
			// attribute as a string and would otherwise interpret it itself,
			// which is how "strict" came to mean Strict on Fiber and Lax on
			// net/http. The four canonical spellings mean the same to both.
			mode := i18n.ResolveCookieSameSite(cfg.CookieSameSite)

			c.Cookie(&fiber.Cookie{
				Name:     cfg.CookieName,
				Value:    string(lang),
				MaxAge:   cfg.CookieMaxAge,
				Path:     cfg.CookiePath,
				Secure:   cfg.CookieSecure || mode.RequiresSecure(),
				HTTPOnly: !cfg.DisableCookieHTTPOnly,
				SameSite: string(mode),
			})
		}

		return c.Next()
	}
}

// SimpleMiddleware creates a middleware that only detects language.
// It is the Fiber counterpart of i18n.SimpleMiddleware.
func SimpleMiddleware() fiber.Handler { return Middleware() }

// Language extracts the detected language from Fiber locals.
func Language(c fiber.Ctx) i18n.Language {
	if lang, ok := c.Locals(i18n.LocalsLanguageKey).(i18n.Language); ok {
		return lang
	}
	return i18n.DefaultLanguage
}

// Bundle extracts the bundle from Fiber locals.
func Bundle(c fiber.Ctx) *i18n.Bundle {
	if bundle, ok := c.Locals(i18n.LocalsBundleKey).(*i18n.Bundle); ok {
		return bundle
	}
	return i18n.DefaultBundle
}

// T returns the translation for key in the request's language.
func T(c fiber.Ctx, key string) string {
	return Bundle(c).GetTranslation(Language(c), key)
}

// Tf is T with printf arguments applied.
//
// Note: the function this replaces, i18n.TfFromFiber, discarded args entirely
// -- it just called TFromFiber -- so "%s" placeholders came out literal on
// Fiber while TfFromContext formatted them correctly. This one follows
// TfFromContext: look the key up, then format, leaving a missing key untouched.
func Tf(c fiber.Ctx, key string, args ...interface{}) string {
	text, found := Bundle(c).LookupTranslation(Language(c), key)
	return i18n.FormatTranslation(text, found, args...)
}
