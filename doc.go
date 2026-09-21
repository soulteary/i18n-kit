// Package i18n provides internationalization for Go applications: translation
// bundles, language lookup and formatting, and the request-language detection
// chain that the adapter subpackages drive.
//
// It depends on nothing outside the standard library, and since v4 it does not
// import net/http either -- translation is just as useful in a CLI printing
// localized help, and net/http costs such a binary well over a hundred packages
// it has no use for. The handlers, the middleware and the *http.Request helpers
// live in the httpadapter subpackage; YAML translation files live in
// yamlloader, because JSON needs only the standard library and most
// translation sets are JSON.
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
// The httpadapter subpackage covers net/http and fiberadapter covers Fiber v3.
// Importing this package links neither; only importing the adapter does. For
// another framework, implement RequestSource and read
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
