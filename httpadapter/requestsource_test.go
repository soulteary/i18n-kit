package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/soulteary/i18n-kit/v4"
)

func TestRequestSourceOf(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh&other=x", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "ja"})
	req.Header.Set("X-Language", "de")

	src := RequestSourceOf(req)

	assert.Equal(t, "zh", src.Query("lang"))
	assert.Equal(t, "ja", src.Cookie("lang"))
	assert.Equal(t, "de", src.Header("X-Language"))
	// Header lookup is canonicalised the way net/http canonicalises it.
	assert.Equal(t, "de", src.Header("x-language"))
}

func TestRequestSourceOf_AbsentValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	src := RequestSourceOf(req)

	assert.Empty(t, src.Query("lang"))
	assert.Empty(t, src.Cookie("lang"))
	assert.Empty(t, src.Header("X-Language"))
}

// A request built by hand rather than by net/http can carry a nil URL; reading
// a query parameter off it must not panic.
func TestRequestSourceOf_NilURL(t *testing.T) {
	req := &http.Request{Header: http.Header{}}

	src := RequestSourceOf(req)

	require.NotPanics(t, func() { assert.Empty(t, src.Query("lang")) })
}

// A cookie whose value is empty is "not found", not an empty language: it has
// to fall through to the next detection method.
func TestRequestSourceOf_EmptyCookieValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Cookie", "lang=")
	req.Header.Set("X-Language", "de")

	assert.Empty(t, RequestSourceOf(req).Cookie("lang"))
	assert.Equal(t, i18n.LangDE, DetectWith(i18n.DefaultDetector, req))
}

// DetectFromRequest is the same chain as Detect, reached through the net/http
// source rather than a hand-written one.
func TestDetectFromRequest_MatchesDetectOverStdSource(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?lang=ko", nil)

	assert.Equal(t,
		i18n.DefaultDetector.Detect(RequestSourceOf(req)),
		DetectWith(i18n.DefaultDetector, req),
	)
}

// The returned i18n.RequestSource reads the request live rather than snapshotting
// it, so an adapter may hold one for the life of a request.
func TestRequestSourceOf_ReadsRequestLive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	src := RequestSourceOf(req)

	require.Empty(t, src.Query("lang"))
	req.URL.RawQuery = url.Values{"lang": {"zh"}}.Encode()

	assert.Equal(t, "zh", src.Query("lang"))
}
