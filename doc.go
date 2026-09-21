// Package i18n provides internationalization for Go applications: translation
// bundles, language detection from HTTP requests, and net/http middleware.
//
// # Framework support
//
// This package depends on no web framework. Detection runs against a
// [RequestSource] -- three string lookups -- so a middleware for any framework
// is a small adapter over the same chain:
//
//	type RequestSource interface {
//		Query(name string) string
//		Cookie(name string) string
//		Header(name string) string
//	}
//
// [StdMiddleware] covers net/http, and the fiberadapter subpackage covers
// Fiber v3. Importing this package does not link Fiber; only importing
// fiberadapter does. For another framework, implement RequestSource and read
// the shared rules from here rather than restating them: [ResolveMiddlewareConfig]
// for config defaults, [ResolveCookieSameSite] for the cookie attribute,
// [FormatTranslation] for the missing-key rule, and [LocalsLanguageKey] /
// [LocalsBundleKey] for where to store the result.
//
// # Missing translations
//
// A translation is used as a printf format string, and a missing key is
// returned as-is with the arguments NOT applied -- see [FormatTranslation].
// Formatting a missing key would append "%!(EXTRA string=...)", putting the
// arguments, often an email address or a user id, into the message shown to
// whoever triggered it.
//
// # Thread safety
//
// [Bundle] and [Translator] are safe for concurrent use, as are the
// package-level functions.
package i18n
