package i18n

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
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

func TestFiberMiddleware_Basic(t *testing.T) {
	app := fiber.New()

	app.Use(FiberMiddleware())

	app.Get("/", func(c fiber.Ctx) error {
		lang := LanguageFromFiberLocals(c)
		return c.SendString(string(lang))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "zh", string(body))
}

func TestFiberMiddleware_WithBundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Fiber Hello")

	app := fiber.New()

	app.Use(FiberMiddleware(MiddlewareConfig{
		Bundle: bundle,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		b := BundleFromFiberLocals(c)
		return c.SendString(b.GetTranslation(LangEN, "greeting"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Fiber Hello", string(body))
}

func TestFiberMiddleware_SetCookie(t *testing.T) {
	app := fiber.New()

	app.Use(FiberMiddleware(MiddlewareConfig{
		SetCookie: true,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Check cookie was set
	cookies := resp.Cookies()
	var langCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "lang" {
			langCookie = c
			break
		}
	}

	require.NotNil(t, langCookie)
	assert.Equal(t, "zh", langCookie.Value)
}

func TestFiberMiddleware_Next(t *testing.T) {
	app := fiber.New()

	app.Use(FiberMiddleware(MiddlewareConfig{
		Next: func(c fiber.Ctx) bool {
			return c.Path() == "/skip"
		},
	}))

	app.Get("/skip", func(c fiber.Ctx) error {
		lang := LanguageFromFiberLocals(c)
		return c.SendString(string(lang))
	})

	app.Get("/normal", func(c fiber.Ctx) error {
		lang := LanguageFromFiberLocals(c)
		return c.SendString(string(lang))
	})

	// Should skip middleware
	req := httptest.NewRequest(http.MethodGet, "/skip?lang=zh", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, string(DefaultLanguage), string(body))

	// Should not skip
	req = httptest.NewRequest(http.MethodGet, "/normal?lang=zh", nil)
	resp, _ = app.Test(req)
	body, _ = io.ReadAll(resp.Body)
	assert.Equal(t, "zh", string(body))
}

func TestTFromFiber(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")

	app := fiber.New()

	app.Use(FiberMiddleware(MiddlewareConfig{
		Bundle: bundle,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(TFromFiber(c, "greeting"))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "你好", string(body))
}

func TestLanguageFromFiberLocals_NoValue(t *testing.T) {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		lang := LanguageFromFiberLocals(c)
		return c.SendString(string(lang))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, string(DefaultLanguage), string(body))
}

func TestBundleFromFiberLocals_NoValue(t *testing.T) {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		bundle := BundleFromFiberLocals(c)
		if bundle == DefaultBundle {
			return c.SendString("default")
		}
		return c.SendString("custom")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "default", string(body))
}

func TestSimpleFiberMiddleware(t *testing.T) {
	app := fiber.New()

	app.Use(SimpleFiberMiddleware())

	app.Get("/", func(c fiber.Ctx) error {
		lang := LanguageFromFiberLocals(c)
		return c.SendString(string(lang))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=ko", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ko", string(body))
}
