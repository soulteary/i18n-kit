package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ResolveCookieSameSite is the one place CookieSameSite is interpreted. Every
// middleware reads it, so the rules are tested here rather than once per
// framework -- which is what let "strict" mean Strict on Fiber and Lax on
// net/http.
func TestResolveCookieSameSite(t *testing.T) {
	tests := []struct {
		value string
		want  CookieSameSiteMode
	}{
		{"Lax", SameSiteLax},
		{"Strict", SameSiteStrict},
		{"None", SameSiteNone},
		{"disabled", SameSiteDisabled},

		// Case-insensitive: a config read from YAML or an env var rarely
		// arrives in the canonical spelling.
		{"lax", SameSiteLax},
		{"strict", SameSiteStrict},
		{"none", SameSiteNone},
		{"DISABLED", SameSiteDisabled},
		{"StRiCt", SameSiteStrict},

		{"  Strict  ", SameSiteStrict},
		{"\tNone\n", SameSiteNone},

		// Unset and unrecognised both mean the default rather than reaching a
		// browser as a malformed attribute.
		{"", SameSiteLax},
		{"   ", SameSiteLax},
		{"bogus", SameSiteLax},
		{"SameSite=Strict", SameSiteLax},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			assert.Equal(t, tt.want, ResolveCookieSameSite(tt.value))
		})
	}
}

// Resolving is idempotent: a mode's own string value resolves back to itself,
// which is what lets an adapter hand the canonical spelling to a framework that
// takes the attribute as a string.
func TestResolveCookieSameSite_IsIdempotent(t *testing.T) {
	for _, mode := range []CookieSameSiteMode{SameSiteLax, SameSiteStrict, SameSiteNone, SameSiteDisabled} {
		assert.Equal(t, mode, ResolveCookieSameSite(string(mode)))
	}
}

func TestCookieSameSiteMode_HTTPSameSite(t *testing.T) {
	tests := []struct {
		mode     CookieSameSiteMode
		want     http.SameSite
		wantWrit bool
	}{
		{SameSiteLax, http.SameSiteLaxMode, true},
		{SameSiteStrict, http.SameSiteStrictMode, true},
		{SameSiteNone, http.SameSiteNoneMode, true},
		{SameSiteDisabled, 0, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			got, write := tt.mode.HTTPSameSite()
			assert.Equal(t, tt.wantWrit, write)
			if write {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// "disabled" means no attribute at all, which is not the same as Lax -- the
// browser's own default applies instead.
func TestCookieSameSiteMode_DisabledWritesNoAttribute(t *testing.T) {
	_, write := SameSiteDisabled.HTTPSameSite()
	assert.False(t, write)

	laxMode, write := SameSiteLax.HTTPSameSite()
	assert.True(t, write)
	assert.Equal(t, http.SameSiteLaxMode, laxMode)
}

// SameSite=None without Secure is rejected by current browsers, so the cookie
// is never stored and detection silently falls back to Accept-Language on every
// request. Only None obliges it.
func TestCookieSameSiteMode_RequiresSecure(t *testing.T) {
	assert.True(t, SameSiteNone.RequiresSecure())
	assert.False(t, SameSiteLax.RequiresSecure())
	assert.False(t, SameSiteStrict.RequiresSecure())
	assert.False(t, SameSiteDisabled.RequiresSecure())
}

// StdMiddleware honours RequiresSecure even when CookieSecure says otherwise:
// the alternative is emitting a cookie the browser throws away.
func TestStdMiddleware_NoneForcesSecure(t *testing.T) {
	handler := StdMiddleware(MiddlewareConfig{
		SetCookie:      true,
		CookieSameSite: "None",
		CookieSecure:   false,
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	rec := newRecordedCookie(t, handler)

	assert.Equal(t, http.SameSiteNoneMode, rec.SameSite)
	assert.True(t, rec.Secure, "SameSite=None without Secure is dropped by browsers")
}

// Nothing else promotes Secure on its own.
func TestStdMiddleware_SecureNotForcedOtherwise(t *testing.T) {
	for _, sameSite := range []string{"", "Lax", "Strict", "disabled"} {
		handler := StdMiddleware(MiddlewareConfig{
			SetCookie:      true,
			CookieSameSite: sameSite,
		})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

		assert.False(t, newRecordedCookie(t, handler).Secure, "SameSite=%q", sameSite)
	}
}

func newRecordedCookie(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))

	for _, c := range rec.Result().Cookies() {
		if c.Name == "lang" {
			return c
		}
	}
	t.Fatal("no lang cookie was written")
	return nil
}
