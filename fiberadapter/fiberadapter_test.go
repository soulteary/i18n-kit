// Fiber adapter tests, moved here from the root package together with the
// middleware they cover. External test package on purpose: they compile only
// against i18n-kit's exported API, which is what an out-of-tree adapter has.
package fiberadapter_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/soulteary/i18n-kit/v4"
	"github.com/soulteary/i18n-kit/v4/fiberadapter"
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

// Detect and DetectWith are the entry points for code that wants the language
// without the middleware -- a handler deciding on its own, say. Neither was
// exercised by the tests that moved here.

func TestDetect(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.Detect(c)))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=ko", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(i18n.LangKO), string(body))
}

func TestDetect_NoSignalGivesDefault(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.Detect(c)))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(i18n.DefaultLanguage), string(body))
}

// DetectWith takes a specific detector, so a custom priority chain has to be
// the one that runs -- here only the header counts, and the query is ignored.
func TestDetectWith_UsesTheGivenDetector(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority:   []string{"header"},
		HeaderName: "X-Locale",
		Default:    i18n.LangIT,
	})

	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.DetectWith(detector, c)))
	})

	withHeader := httptest.NewRequest(http.MethodGet, "/?lang=ko", nil)
	withHeader.Header.Set("X-Locale", "de")
	resp, err := app.Test(withHeader)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(i18n.LangDE), string(body))

	// The query is not in this detector's chain, so it falls to the default.
	resp2, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=ko", nil))
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	assert.Equal(t, string(i18n.LangIT), string(body2))
}

// Source is the three-method view the root package detects against. Only Query
// was ever exercised; Cookie and Header are how a Fiber request carries a
// language in every deployment that does not put it in the URL.
func TestSource_Lookups(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error {
		src := fiberadapter.Source{C: c}
		return c.JSON(map[string]string{
			"query":  src.Query("lang"),
			"cookie": src.Cookie("lang"),
			"header": src.Header("X-Language"),
			"accept": src.Header("Accept-Language"),
			"absent": src.Query("nope") + src.Cookie("nope") + src.Header("X-Nope"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "ja"})
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "fr;q=0.9")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var got map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, "zh", got["query"])
	assert.Equal(t, "ja", got["cookie"])
	assert.Equal(t, "de", got["header"])
	assert.Equal(t, "fr;q=0.9", got["accept"])
	assert.Empty(t, got["absent"])
}

// The embedded i18n.MiddlewareConfig has to actually reach the detector: a
// Config whose Detector is ignored would silently fall back to the default
// chain, which is the failure mode hardest to notice.
func TestMiddleware_UsesConfiguredDetector(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority:   []string{"cookie"},
		CookieName: "site_lang",
		Default:    i18n.LangPT,
	})

	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{Detector: detector},
	}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.Language(c)))
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	req.AddCookie(&http.Cookie{Name: "site_lang", Value: "ru"})
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(i18n.LangRU), string(body))
}

// Cookie attributes come from the embedded config. Only SetCookie itself was
// covered, so a field dropped on the way through Config would not have shown up.
func TestMiddleware_CookieAttributes(t *testing.T) {
	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{
			SetCookie:    true,
			CookieName:   "site_lang",
			CookieMaxAge: 60,
			CookiePath:   "/app",
			CookieSecure: true,
		},
	}))
	app.Get("/", func(c fiber.Ctx) error { return c.SendString("ok") })

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "site_lang" {
			cookie = c
			break
		}
	}
	require.NotNil(t, cookie, "the configured cookie name should be the one that is set")
	assert.Equal(t, "zh", cookie.Value)
	assert.Equal(t, 60, cookie.MaxAge)
	assert.Equal(t, "/app", cookie.Path)
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
}

// No config at all must still detect: Middleware()'s variadic path resolves to
// the package defaults rather than a zero MiddlewareConfig with a nil detector.
func TestMiddleware_NoConfigUsesDefaults(t *testing.T) {
	app := fiber.New()
	app.Use(fiberadapter.Middleware())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.Language(c)))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, string(i18n.LangRU), string(body))
}

// SetCookie defaults to off; a middleware that started writing cookies on its
// own would be a surprise in a deployment that never asked for one.
func TestMiddleware_DoesNotSetCookieByDefault(t *testing.T) {
	app := fiber.New()
	app.Use(fiberadapter.Middleware())
	app.Get("/", func(c fiber.Ctx) error { return c.SendString("ok") })

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Empty(t, resp.Cookies())
}

// Without the middleware there is nothing in locals, and Tf has to fall back to
// the default bundle and language rather than panic on the missing values.
func TestTf_WithoutMiddleware(t *testing.T) {
	app := fiber.New()
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

// A present translation is still run through the formatter with no arguments,
// so its own printf escapes render -- the rule i18n.FormatTranslation states.
func TestTf_FormatsEscapesWithNoArgs(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	bundle.AddTranslation(i18n.LangEN, "discount", "Save 10%%")

	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{Bundle: bundle},
	}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(fiberadapter.Tf(c, "discount"))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Save 10%", string(body))
}

// T and Tf must read the same bundle and language, so a key with no arguments
// comes out the same through either.
func TestT_AndTf_AgreeWithoutArgs(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	bundle.AddTranslation(i18n.LangZH, "greeting", "你好")

	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{Bundle: bundle},
	}))
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(fiberadapter.T(c, "greeting") + "|" + fiberadapter.Tf(c, "greeting"))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "你好|你好", string(body))
}

// Fiber locals and net/http context values are separate stores, and the keys
// are not interchangeable even though they spell the same string: the root
// package's context keys have an unexported type, while LocalsLanguageKey is an
// untyped string. So the language this middleware puts in locals does NOT read
// back through i18n.LanguageFromContext, and i18n.TFromContext under Fiber
// silently answers in the default language rather than failing.
//
// Pinned because the alternative is finding out in production. Read the
// language with fiberadapter.Language(c); reach for the context helpers only
// after putting the value there yourself.
func TestLocalsAreNotNetHTTPContextValues(t *testing.T) {
	app := fiber.New()
	app.Use(fiberadapter.Middleware())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.Language(c)) + "|" + string(i18n.LanguageFromContext(c.Context())))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "zh|"+string(i18n.DefaultLanguage), string(body),
		"locals carry the detected language; the net/http context does not see it")
}

// The exported keys are what an out-of-tree adapter writes under so that
// fiberadapter.Language and its own reader agree. Their values are therefore
// part of the API, not an implementation detail.
func TestLocalsKeysAreTheOnesTheAdapterWrites(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)

	app := fiber.New()
	app.Use(fiberadapter.Middleware(fiberadapter.Config{
		MiddlewareConfig: i18n.MiddlewareConfig{Bundle: bundle},
	}))
	app.Get("/", func(c fiber.Ctx) error {
		lang, langOK := c.Locals(i18n.LocalsLanguageKey).(i18n.Language)
		got, bundleOK := c.Locals(i18n.LocalsBundleKey).(*i18n.Bundle)
		assert.True(t, langOK)
		assert.True(t, bundleOK)
		assert.Equal(t, i18n.LangZH, lang)
		assert.Same(t, bundle, got)
		return c.SendString("ok")
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/?lang=zh", nil))
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}
