package i18n

import (
	"net/http"
	"strings"
)

// CookieSameSiteMode is the normalised form of MiddlewareConfig.CookieSameSite.
//
// The string form exists because it has to cross frameworks that spell the
// attribute differently: net/http wants an http.SameSite constant, Fiber wants
// a string. Normalising to one of these four first is what keeps the two from
// answering differently -- they used to, because each middleware interpreted
// the raw string itself.
type CookieSameSiteMode string

const (
	// SameSiteLax is the default: the attribute a language cookie wants.
	SameSiteLax CookieSameSiteMode = "Lax"
	// SameSiteStrict never sends the cookie on a cross-site request.
	SameSiteStrict CookieSameSiteMode = "Strict"
	// SameSiteNone sends it on cross-site requests, and requires Secure --
	// see RequiresSecure.
	SameSiteNone CookieSameSiteMode = "None"
	// SameSiteDisabled writes no SameSite attribute at all, leaving the
	// browser's own default to apply.
	SameSiteDisabled CookieSameSiteMode = "disabled"
)

// ResolveCookieSameSite normalises a MiddlewareConfig.CookieSameSite value.
//
// Matching is case-insensitive, an empty value means the default, and anything
// unrecognised falls back to SameSiteLax rather than reaching a browser as a
// malformed attribute.
//
// Framework adapters call this rather than interpreting the string themselves,
// so "strict" and "Strict" mean the same thing everywhere.
func ResolveCookieSameSite(value string) CookieSameSiteMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return SameSiteStrict
	case "none":
		return SameSiteNone
	case "disabled":
		return SameSiteDisabled
	default:
		return SameSiteLax
	}
}

// HTTPSameSite returns the net/http constant for m. The second result is false
// when no SameSite attribute should be written at all, which is not the same as
// writing SameSite=Lax.
func (m CookieSameSiteMode) HTTPSameSite() (http.SameSite, bool) {
	switch m {
	case SameSiteStrict:
		return http.SameSiteStrictMode, true
	case SameSiteNone:
		return http.SameSiteNoneMode, true
	case SameSiteDisabled:
		return 0, false
	default:
		return http.SameSiteLaxMode, true
	}
}

// RequiresSecure reports whether m obliges the cookie to carry Secure.
//
// SameSite=None without Secure is rejected outright by current browsers, so a
// language cookie written that way is silently never stored and detection falls
// back to the Accept-Language header on every request. Middlewares therefore
// set Secure when this returns true, whatever CookieSecure says.
func (m CookieSameSiteMode) RequiresSecure() bool { return m == SameSiteNone }
