package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/soulteary/i18n-kit/v4"
)

func TestStdMiddleware_Basic(t *testing.T) {
	defer i18n.DefaultBundle.Clear()
	i18n.DefaultBundle.AddTranslation(i18n.LangEN, "greeting", "Hello")
	i18n.DefaultBundle.AddTranslation(i18n.LangZH, "greeting", "你好")

	middleware := Middleware()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := Language(r)
		translation := T(r, "greeting")
		_, _ = w.Write([]byte(translation + " (" + string(lang) + ")"))
	}))

	// Test with query param
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "你好 (zh)", rec.Body.String())
}

func TestStdMiddleware_WithBundle(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	bundle.AddTranslation(i18n.LangEN, "greeting", "Custom Hello")

	middleware := Middleware(Config{MiddlewareConfig: i18n.MiddlewareConfig{
		Bundle: bundle,
	}})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := i18n.BundleFromContext(r.Context())
		translation := b.GetTranslation(i18n.LangEN, "greeting")
		_, _ = w.Write([]byte(translation))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "Custom Hello", rec.Body.String())
}

func TestStdMiddleware_SetCookie(t *testing.T) {
	middleware := Middleware(Config{MiddlewareConfig: i18n.MiddlewareConfig{
		SetCookie: true,
	}})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Check cookie was set
	cookies := rec.Result().Cookies()
	var langCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "lang" {
			langCookie = c
			break
		}
	}

	require.NotNil(t, langCookie)
	assert.Equal(t, "zh", langCookie.Value)
	assert.Equal(t, "/", langCookie.Path)
	assert.True(t, langCookie.HttpOnly)
}

func TestStdMiddleware_SetCookieSameSite(t *testing.T) {
	tests := []struct {
		sameSite string
		expected http.SameSite
	}{
		{"Lax", http.SameSiteLaxMode},
		{"Strict", http.SameSiteStrictMode},
		{"None", http.SameSiteNoneMode},
	}

	for _, tt := range tests {
		t.Run(tt.sameSite, func(t *testing.T) {
			middleware := Middleware(Config{MiddlewareConfig: i18n.MiddlewareConfig{
				SetCookie:      true,
				CookieSameSite: tt.sameSite,
			}})

			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("ok"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			cookies := rec.Result().Cookies()
			var langCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "lang" {
					langCookie = c
					break
				}
			}

			require.NotNil(t, langCookie)
			assert.Equal(t, tt.expected, langCookie.SameSite)
		})
	}
}

func TestStdMiddleware_NextStd(t *testing.T) {
	middleware := Middleware(Config{Next: func(r *http.Request) bool {
		return r.URL.Path == "/skip"
	}, MiddlewareConfig: i18n.MiddlewareConfig{}})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := Language(r)
		_, _ = w.Write([]byte(string(lang)))
	}))

	// Should skip middleware
	req := httptest.NewRequest(http.MethodGet, "/skip?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, string(i18n.DefaultLanguage), rec.Body.String())

	// Should not skip
	req = httptest.NewRequest(http.MethodGet, "/normal?lang=zh", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, "zh", rec.Body.String())
}

func TestStdMiddlewareFunc(t *testing.T) {
	middleware := MiddlewareFunc()

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		lang := Language(r)
		_, _ = w.Write([]byte(string(lang)))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=fr", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(t, "fr", rec.Body.String())
}

func TestSimpleMiddleware(t *testing.T) {
	middleware := SimpleMiddleware()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := Language(r)
		_, _ = w.Write([]byte(string(lang)))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "de", rec.Body.String())
}

// i18n.ResolveMiddlewareConfig is what a framework adapter calls instead of
// restating the merge rules, so the rules need testing directly rather than
// only through whichever middleware happens to exercise them.

// Only the second argument onwards is ignored -- the variadic is "zero or one
// config" in practice, and an adapter passing a slice must not get a silent
// merge of both.
func TestStdMiddleware_UsesResolvedConfig(t *testing.T) {
	cfg := i18n.ResolveMiddlewareConfig(i18n.MiddlewareConfig{SetCookie: true, CookieName: "site_lang"})

	handler := Middleware(Config{MiddlewareConfig: i18n.MiddlewareConfig{SetCookie: true, CookieName: "site_lang"}})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, cfg.CookieName, cookies[0].Name)
	assert.Equal(t, cfg.CookieMaxAge, cookies[0].MaxAge)
	assert.Equal(t, cfg.CookiePath, cookies[0].Path)
	assert.Equal(t, !cfg.DisableCookieHTTPOnly, cookies[0].HttpOnly)
}
