package i18n

import (
	"context"
	"fmt"
	"net/http"
)

// contextKey is used to store language in context.
type contextKey string

const languageContextKey contextKey = "i18n-language"

// ContextWithLanguage returns a new context with the language set.
func ContextWithLanguage(ctx context.Context, lang Language) context.Context {
	return context.WithValue(ctx, languageContextKey, lang)
}

// LanguageFromContext extracts the language from the context.
// Returns DefaultLanguage if not found.
func LanguageFromContext(ctx context.Context) Language {
	if ctx == nil {
		return DefaultLanguage
	}
	if lang, ok := ctx.Value(languageContextKey).(Language); ok {
		return lang
	}
	return DefaultLanguage
}

// LanguageFromContextOK extracts the language from the context.
// Returns (Language, false) if not found.
func LanguageFromContextOK(ctx context.Context) (Language, bool) {
	if ctx == nil {
		return "", false
	}
	lang, ok := ctx.Value(languageContextKey).(Language)
	return lang, ok
}

// SetLanguageInRequest sets the language in the request context.
func SetLanguageInRequest(r *http.Request, lang Language) *http.Request {
	return r.WithContext(ContextWithLanguage(r.Context(), lang))
}

// LanguageFromRequest extracts the language from the request context.
// Returns DefaultLanguage if not found.
func LanguageFromRequest(r *http.Request) Language {
	if r == nil {
		return DefaultLanguage
	}
	return LanguageFromContext(r.Context())
}

// TFromContext returns the translated string using the language from context.
func TFromContext(ctx context.Context, key string) string {
	lang := LanguageFromContext(ctx)
	return DefaultBundle.GetTranslation(lang, key)
}

// TfFromContext returns a formatted translated string using the language from context.
func TfFromContext(ctx context.Context, key string, args ...interface{}) string {
	return fmt.Sprintf(TFromContext(ctx, key), args...)
}

// TFromRequest returns the translated string using the language from request context.
func TFromRequest(r *http.Request, key string) string {
	lang := LanguageFromRequest(r)
	return DefaultBundle.GetTranslation(lang, key)
}

// TfFromRequest returns a formatted translated string using the language from request context.
func TfFromRequest(r *http.Request, key string, args ...interface{}) string {
	return fmt.Sprintf(TFromRequest(r, key), args...)
}

// BundleContextKey is the context key for storing a custom bundle.
const bundleContextKey contextKey = "i18n-bundle"

// ContextWithBundle returns a new context with a custom bundle.
func ContextWithBundle(ctx context.Context, bundle *Bundle) context.Context {
	return context.WithValue(ctx, bundleContextKey, bundle)
}

// BundleFromContext extracts the bundle from the context.
// Returns DefaultBundle if not found.
func BundleFromContext(ctx context.Context) *Bundle {
	if ctx == nil {
		return DefaultBundle
	}
	if bundle, ok := ctx.Value(bundleContextKey).(*Bundle); ok {
		return bundle
	}
	return DefaultBundle
}

// TFromContextWithBundle returns the translated string using the bundle and language from context.
func TFromContextWithBundle(ctx context.Context, key string) string {
	bundle := BundleFromContext(ctx)
	lang := LanguageFromContext(ctx)
	return bundle.GetTranslation(lang, key)
}

// TfFromContextWithBundle returns a formatted translated string using the bundle and language from context.
func TfFromContextWithBundle(ctx context.Context, key string, args ...interface{}) string {
	return fmt.Sprintf(TFromContextWithBundle(ctx, key), args...)
}
