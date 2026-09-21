package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	i18n "github.com/soulteary/i18n-kit/v4"
)

func TestCookieSameSiteMode_HTTPSameSite(t *testing.T) {
	tests := []struct {
		mode     i18n.CookieSameSiteMode
		want     http.SameSite
		wantWrit bool
	}{
		{i18n.SameSiteLax, http.SameSiteLaxMode, true},
		{i18n.SameSiteStrict, http.SameSiteStrictMode, true},
		{i18n.SameSiteNone, http.SameSiteNoneMode, true},
		{i18n.SameSiteDisabled, 0, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			got, write := SameSite(tt.mode)
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
	_, write := SameSite(i18n.SameSiteDisabled)
	assert.False(t, write)

	laxMode, write := SameSite(i18n.SameSiteLax)
	assert.True(t, write)
	assert.Equal(t, http.SameSiteLaxMode, laxMode)
}

// SameSite=None without Secure is rejected by current browsers, so the cookie
// is never stored and detection silently falls back to Accept-i18n.Language on every
// request. Only None obliges it.
func TestStdMiddleware_NoneForcesSecure(t *testing.T) {
	handler := Middleware(Config{MiddlewareConfig: i18n.MiddlewareConfig{
		SetCookie:      true,
		CookieSameSite: "None",
		CookieSecure:   false,
	}})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	rec := newRecordedCookie(t, handler)

	assert.Equal(t, http.SameSiteNoneMode, rec.SameSite)
	assert.True(t, rec.Secure, "SameSite=None without Secure is dropped by browsers")
}

// Nothing else promotes Secure on its own.
func TestStdMiddleware_SecureNotForcedOtherwise(t *testing.T) {
	for _, sameSite := range []string{"", "Lax", "Strict", "disabled"} {
		handler := Middleware(Config{MiddlewareConfig: i18n.MiddlewareConfig{
			SetCookie:      true,
			CookieSameSite: sameSite,
		}})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

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
