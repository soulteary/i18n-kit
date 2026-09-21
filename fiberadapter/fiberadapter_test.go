// Fiber adapter tests, moved here from the root package together with the
// middleware they cover. External test package on purpose: they compile only
// against i18n-kit's exported API, which is what an out-of-tree adapter has.
package fiberadapter_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/soulteary/i18n-kit/v2"
	"github.com/soulteary/i18n-kit/v2/fiberadapter"
)

func TestFiberMiddleware_Basic(t *testing.T) {
	app := fiber.New()

	app.Use(fiberadapter.Middleware())

	app.Get("/", func(c fiber.Ctx) error {
		lang := fiberadapter.Language(c)
		return c.SendString(string(lang))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "zh", string(body))
}

func TestFiberMiddleware_WithBundle(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	bundle.AddTranslation(i18n.LangEN, "greeting", "Fiber Hello")

	app := fiber.New()

	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{
			Bundle: bundle,
		},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		b := fiberadapter.Bundle(c)
		return c.SendString(b.GetTranslation(i18n.LangEN, "greeting"))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Fiber Hello", string(body))
}

func TestFiberMiddleware_SetCookie(t *testing.T) {
	app := fiber.New()

	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{
			SetCookie: true,
		},
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

	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		Next: func(c fiber.Ctx) bool {
			return c.Path() == "/skip"
		},
	}))

	app.Get("/skip", func(c fiber.Ctx) error {
		lang := fiberadapter.Language(c)
		return c.SendString(string(lang))
	})

	app.Get("/normal", func(c fiber.Ctx) error {
		lang := fiberadapter.Language(c)
		return c.SendString(string(lang))
	})

	// Should skip middleware
	req := httptest.NewRequest(http.MethodGet, "/skip?lang=zh", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, string(i18n.DefaultLanguage), string(body))

	// Should not skip
	req = httptest.NewRequest(http.MethodGet, "/normal?lang=zh", nil)
	resp, _ = app.Test(req)
	body, _ = io.ReadAll(resp.Body)
	assert.Equal(t, "zh", string(body))
}

func TestTFromFiber(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	bundle.AddTranslation(i18n.LangEN, "greeting", "Hello")
	bundle.AddTranslation(i18n.LangZH, "greeting", "你好")

	app := fiber.New()

	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{
			Bundle: bundle,
		},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(fiberadapter.T(c, "greeting"))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "你好", string(body))
}

func TestLanguageFromFiberLocals_NoValue(t *testing.T) {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		lang := fiberadapter.Language(c)
		return c.SendString(string(lang))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, string(i18n.DefaultLanguage), string(body))
}

func TestBundleFromFiberLocals_NoValue(t *testing.T) {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		bundle := fiberadapter.Bundle(c)
		if bundle == i18n.DefaultBundle {
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

	app.Use(fiberadapter.SimpleMiddleware())

	app.Get("/", func(c fiber.Ctx) error {
		lang := fiberadapter.Language(c)
		return c.SendString(string(lang))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=ko", nil)
	resp, _ := app.Test(req)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ko", string(body))
}

// Tf applies its printf arguments. The function it replaces, i18n.TfFromFiber,
// discarded them -- it just delegated to TFromFiber -- so a "%s" placeholder
// came out literal on Fiber while i18n.TfFromContext formatted it correctly.
// This pins the fix.
func TestTfAppliesArgs(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	bundle.AddTranslation(i18n.LangZH, "hello", "你好，%s")

	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{Bundle: bundle},
	}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(fiberadapter.Tf(c, "hello", "世界"))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "你好，世界", string(body))
}

// A key with no translation is returned as-is and the arguments are NOT
// applied -- the same rule i18n.TfFromContext follows, so the two agree.
func TestTfMissingKeyIsReturnedUnformatted(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)

	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{Bundle: bundle},
	}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(fiberadapter.Tf(c, "no.such.key", "世界"))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "no.such.key", string(body))
}
