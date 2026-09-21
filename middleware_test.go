package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	assert.Equal(t, DefaultDetector, config.Detector)
	assert.Nil(t, config.Bundle)
	assert.False(t, config.SetCookie)
	assert.Equal(t, "lang", config.CookieName)
	assert.Equal(t, 86400*365, config.CookieMaxAge)
	assert.Equal(t, "/", config.CookiePath)
	assert.False(t, config.CookieSecure)
	assert.True(t, config.CookieHTTPOnly)
	assert.Equal(t, "Lax", config.CookieSameSite)
}

func TestStdMiddleware_Basic(t *testing.T) {
	defer DefaultBundle.Clear()
	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello")
	DefaultBundle.AddTranslation(LangZH, "greeting", "你好")

	middleware := StdMiddleware()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := LanguageFromRequest(r)
		translation := TFromRequest(r, "greeting")
		_, _ = w.Write([]byte(translation + " (" + string(lang) + ")"))
	}))

	// Test with query param
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "你好 (zh)", rec.Body.String())
}

func TestStdMiddleware_WithBundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Custom Hello")

	middleware := StdMiddleware(MiddlewareConfig{
		Bundle: bundle,
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := BundleFromContext(r.Context())
		translation := b.GetTranslation(LangEN, "greeting")
		_, _ = w.Write([]byte(translation))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "Custom Hello", rec.Body.String())
}

func TestStdMiddleware_SetCookie(t *testing.T) {
	middleware := StdMiddleware(MiddlewareConfig{
		SetCookie: true,
	})

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
			middleware := StdMiddleware(MiddlewareConfig{
				SetCookie:      true,
				CookieSameSite: tt.sameSite,
			})

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
	middleware := StdMiddleware(MiddlewareConfig{
		NextStd: func(r *http.Request) bool {
			return r.URL.Path == "/skip"
		},
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := LanguageFromRequest(r)
		_, _ = w.Write([]byte(string(lang)))
	}))

	// Should skip middleware
	req := httptest.NewRequest(http.MethodGet, "/skip?lang=zh", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, string(DefaultLanguage), rec.Body.String())

	// Should not skip
	req = httptest.NewRequest(http.MethodGet, "/normal?lang=zh", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, "zh", rec.Body.String())
}

func TestStdMiddlewareFunc(t *testing.T) {
	middleware := StdMiddlewareFunc()

	handler := middleware(func(w http.ResponseWriter, r *http.Request) {
		lang := LanguageFromRequest(r)
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
		lang := LanguageFromRequest(r)
		_, _ = w.Write([]byte(string(lang)))
	}))

	req := httptest.NewRequest(http.MethodGet, "/?lang=de", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "de", rec.Body.String())
}

// ResolveMiddlewareConfig is what a framework adapter calls instead of
// restating the merge rules, so the rules need testing directly rather than
// only through whichever middleware happens to exercise them.

func TestResolveMiddlewareConfig_NoArgsGivesDefaults(t *testing.T) {
	assert.Equal(t, DefaultMiddlewareConfig(), ResolveMiddlewareConfig())
}

func TestResolveMiddlewareConfig_ZeroConfigKeepsDefaults(t *testing.T) {
	cfg := ResolveMiddlewareConfig(MiddlewareConfig{})

	assert.Equal(t, DefaultDetector, cfg.Detector)
	assert.Nil(t, cfg.Bundle)
	assert.Equal(t, "lang", cfg.CookieName)
	assert.Equal(t, 86400*365, cfg.CookieMaxAge)
	assert.Equal(t, "/", cfg.CookiePath)
	assert.Equal(t, "Lax", cfg.CookieSameSite)
	assert.True(t, cfg.CookieHTTPOnly)
	assert.False(t, cfg.SetCookie)
	assert.False(t, cfg.CookieSecure)
}

func TestResolveMiddlewareConfig_UserValuesOverrideDefaults(t *testing.T) {
	detector := NewDetector(DetectorConfig{Priority: []string{"header"}})
	bundle := NewBundle(LangEN)
	nextStd := func(*http.Request) bool { return true }

	cfg := ResolveMiddlewareConfig(MiddlewareConfig{
		Detector:       detector,
		Bundle:         bundle,
		SetCookie:      true,
		CookieName:     "site_lang",
		CookieMaxAge:   60,
		CookiePath:     "/app",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Strict",
		NextStd:        nextStd,
	})

	assert.Same(t, detector, cfg.Detector)
	assert.Same(t, bundle, cfg.Bundle)
	assert.True(t, cfg.SetCookie)
	assert.Equal(t, "site_lang", cfg.CookieName)
	assert.Equal(t, 60, cfg.CookieMaxAge)
	assert.Equal(t, "/app", cfg.CookiePath)
	assert.True(t, cfg.CookieSecure)
	assert.True(t, cfg.CookieHTTPOnly)
	assert.Equal(t, "Strict", cfg.CookieSameSite)
	require.NotNil(t, cfg.NextStd)
	assert.True(t, cfg.NextStd(nil))
}

// Only the second argument onwards is ignored -- the variadic is "zero or one
// config" in practice, and an adapter passing a slice must not get a silent
// merge of both.
func TestResolveMiddlewareConfig_IgnoresExtraConfigs(t *testing.T) {
	cfg := ResolveMiddlewareConfig(
		MiddlewareConfig{CookieName: "first"},
		MiddlewareConfig{CookieName: "second", CookiePath: "/ignored"},
	)

	assert.Equal(t, "first", cfg.CookieName)
	assert.Equal(t, "/", cfg.CookiePath)
}

// CookieHTTPOnly defaults to true and is only taken from the caller once the
// caller has shown it is managing cookies, which mergeConfig infers from
// CookieName or CookieSameSite being set. The upshot is a trap: naming the
// cookie and nothing else silently turns HttpOnly off. Pinned so the rule the
// adapters now share is explicit rather than folklore.
func TestResolveMiddlewareConfig_CookieHTTPOnlyRule(t *testing.T) {
	tests := []struct {
		name string
		user MiddlewareConfig
		want bool
	}{
		{"untouched keeps the secure default", MiddlewareConfig{}, true},
		{"unrelated field keeps the default", MiddlewareConfig{CookieMaxAge: 60}, true},
		{"naming the cookie hands control over", MiddlewareConfig{CookieName: "site_lang"}, false},
		{"setting SameSite hands control over", MiddlewareConfig{CookieSameSite: "Strict"}, false},
		{"and can then be asked for explicitly", MiddlewareConfig{CookieName: "site_lang", CookieHTTPOnly: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ResolveMiddlewareConfig(tt.user).CookieHTTPOnly)
		})
	}
}

// StdMiddleware must go through ResolveMiddlewareConfig rather than keeping its
// own copy of the defaults: that equivalence is the whole point of exporting it.
func TestStdMiddleware_UsesResolvedConfig(t *testing.T) {
	cfg := ResolveMiddlewareConfig(MiddlewareConfig{SetCookie: true, CookieName: "site_lang", CookieHTTPOnly: true})

	handler := StdMiddleware(MiddlewareConfig{SetCookie: true, CookieName: "site_lang", CookieHTTPOnly: true})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, cfg.CookieName, cookies[0].Name)
	assert.Equal(t, cfg.CookieMaxAge, cookies[0].MaxAge)
	assert.Equal(t, cfg.CookiePath, cookies[0].Path)
	assert.Equal(t, cfg.CookieHTTPOnly, cookies[0].HttpOnly)
}
