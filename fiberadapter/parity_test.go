// Cross-framework parity. The refactor's whole claim is that detection now has
// one implementation and the adapters only say where the three values come
// from, so net/http and Fiber cannot answer differently. That claim is what
// these tests check: the same request, through both middlewares, has to yield
// the same language. Before the refactor a method added to one chain and
// forgotten in the other would have surfaced only as "it works on net/http but
// not on Fiber"; now it surfaces here.
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

// shape mutates a fresh GET / request; each case is applied to both stacks.
type shape func(*http.Request)

func newRequest(s shape) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s(req)
	return req
}

func langViaNetHTTP(t *testing.T, s shape) string {
	t.Helper()

	var got i18n.Language
	handler := i18n.StdMiddleware()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = i18n.LanguageFromRequest(r)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), newRequest(s))

	return string(got)
}

func langViaFiber(t *testing.T, s shape) string {
	t.Helper()

	app := fiber.New()
	app.Use(fiberadapter.Middleware())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(string(fiberadapter.Language(c)))
	})

	resp, err := app.Test(newRequest(s))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}

func query(raw string) shape {
	return func(r *http.Request) { r.URL.RawQuery = raw }
}

func cookie(name, value string) shape {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: name, Value: value}) }
}

func header(name, value string) shape {
	return func(r *http.Request) { r.Header.Set(name, value) }
}

func all(shapes ...shape) shape {
	return func(r *http.Request) {
		for _, s := range shapes {
			s(r)
		}
	}
}

func TestDetectionParityWithNetHTTP(t *testing.T) {
	tests := []struct {
		name  string
		shape shape
		want  i18n.Language
	}{
		{
			name:  "no signal at all",
			shape: func(*http.Request) {},
			want:  i18n.DefaultLanguage,
		},
		{
			name:  "query parameter",
			shape: query("lang=zh"),
			want:  i18n.LangZH,
		},
		{
			name:  "cookie",
			shape: cookie("lang", "ja"),
			want:  i18n.LangJA,
		},
		{
			name:  "X-Language header",
			shape: header("X-Language", "de"),
			want:  i18n.LangDE,
		},
		{
			name:  "Accept-Language header",
			shape: header("Accept-Language", "fr-FR,fr;q=0.9,en;q=0.8"),
			want:  i18n.LangFR,
		},
		{
			name:  "Accept-Language obeys q-values rather than order",
			shape: header("Accept-Language", "en;q=0.3,ko;q=0.9"),
			want:  i18n.LangKO,
		},
		{
			name: "query outranks cookie, header and Accept-Language",
			shape: all(
				query("lang=zh"),
				cookie("lang", "ja"),
				header("X-Language", "de"),
				header("Accept-Language", "fr"),
			),
			want: i18n.LangZH,
		},
		{
			name: "cookie outranks header and Accept-Language",
			shape: all(
				cookie("lang", "ja"),
				header("X-Language", "de"),
				header("Accept-Language", "fr"),
			),
			want: i18n.LangJA,
		},
		{
			name: "header outranks Accept-Language",
			shape: all(
				header("X-Language", "de"),
				header("Accept-Language", "fr"),
			),
			want: i18n.LangDE,
		},
		{
			name:  "an unrecognised query value falls through to the cookie",
			shape: all(query("lang=klingon"), cookie("lang", "ja")),
			want:  i18n.LangJA,
		},
		{
			name:  "an empty query value falls through to the cookie",
			shape: all(query("lang="), cookie("lang", "ja")),
			want:  i18n.LangJA,
		},
		{
			name:  "an empty cookie value falls through to the header",
			shape: all(header("Cookie", "lang="), header("X-Language", "de")),
			want:  i18n.LangDE,
		},
		{
			name:  "a regional variant normalises to its base language",
			shape: query("lang=zh-TW"),
			want:  i18n.LangZH,
		},
		{
			name:  "an underscore variant normalises too",
			shape: cookie("lang", "pt_BR"),
			want:  i18n.LangPT,
		},
		{
			name:  "an unsupported language everywhere falls back to the default",
			shape: all(query("lang=klingon"), cookie("lang", "elvish"), header("Accept-Language", "sindarin")),
			want:  i18n.DefaultLanguage,
		},
		{
			name:  "a wildcard Accept-Language is not a language",
			shape: header("Accept-Language", "*"),
			want:  i18n.DefaultLanguage,
		},
		{
			name:  "the query parameter is matched by name, not position",
			shape: query("other=zh&lang=ko"),
			want:  i18n.LangKO,
		},
		{
			name:  "a repeated query parameter takes the first value",
			shape: query("lang=ko&lang=zh"),
			want:  i18n.LangKO,
		},
		{
			name:  "header lookup is case-insensitive",
			shape: header("x-language", "it"),
			want:  i18n.LangIT,
		},
		{
			name:  "an unsupported Accept-Language entry yields to a supported one",
			shape: header("Accept-Language", "sindarin;q=1.0,es;q=0.4"),
			want:  i18n.LangES,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std := langViaNetHTTP(t, tt.shape)
			fbr := langViaFiber(t, tt.shape)

			assert.Equal(t, string(tt.want), std, "net/http")
			assert.Equal(t, std, fbr, "Fiber disagrees with net/http")
		})
	}
}

// The convenience detectors take the same route as the middlewares, so they
// have to agree with them and with each other.
func TestDetectParityWithDetectFromRequest(t *testing.T) {
	shapes := []shape{
		query("lang=zh"),
		cookie("lang", "ja"),
		header("X-Language", "de"),
		header("Accept-Language", "fr;q=0.9"),
		func(*http.Request) {},
	}

	for _, s := range shapes {
		app := fiber.New()
		app.Get("/", func(c fiber.Ctx) error {
			return c.SendString(string(fiberadapter.Detect(c)))
		})

		resp, err := app.Test(newRequest(s))
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		assert.Equal(t, string(i18n.DetectFromRequest(newRequest(s))), string(body))
	}
}
