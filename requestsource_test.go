package i18n

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSource is what an out-of-tree adapter (Echo, Gin, chi) looks like: three
// string lookups and nothing else. The chain is driven through it on purpose --
// these tests would still compile and pass if both in-tree adapters were
// deleted, which is the claim RequestSource makes.
//
// It also counts lookups, so a test can assert which sources the chain touched
// rather than only which language came out.
type fakeSource struct {
	query   map[string]string
	cookie  map[string]string
	header  map[string]string
	queries []string
	cookies []string
	headers []string
}

func (s *fakeSource) Query(name string) string {
	s.queries = append(s.queries, name)
	return s.query[name]
}

func (s *fakeSource) Cookie(name string) string {
	s.cookies = append(s.cookies, name)
	return s.cookie[name]
}

func (s *fakeSource) Header(name string) string {
	s.headers = append(s.headers, name)
	return s.header[name]
}

func TestDetect_CustomRequestSource(t *testing.T) {
	tests := []struct {
		name   string
		config DetectorConfig
		src    fakeSource
		want   Language
	}{
		{
			name:   "query",
			config: DetectorConfig{},
			src:    fakeSource{query: map[string]string{"lang": "zh"}},
			want:   LangZH,
		},
		{
			name:   "cookie",
			config: DetectorConfig{},
			src:    fakeSource{cookie: map[string]string{"lang": "ja"}},
			want:   LangJA,
		},
		{
			name:   "header",
			config: DetectorConfig{},
			src:    fakeSource{header: map[string]string{"X-Language": "de"}},
			want:   LangDE,
		},
		{
			name:   "accept-language",
			config: DetectorConfig{AcceptLanguage: true},
			src:    fakeSource{header: map[string]string{"Accept-Language": "fr-FR,fr;q=0.9"}},
			want:   LangFR,
		},
		{
			name:   "nothing set falls back to the default",
			config: DetectorConfig{Default: LangIT},
			src:    fakeSource{},
			want:   LangIT,
		},
		{
			name:   "query wins over everything behind it",
			config: DetectorConfig{AcceptLanguage: true},
			src: fakeSource{
				query:  map[string]string{"lang": "zh"},
				cookie: map[string]string{"lang": "ja"},
				header: map[string]string{"X-Language": "de", "Accept-Language": "fr"},
			},
			want: LangZH,
		},
		{
			name:   "an unrecognised value falls through to the next method",
			config: DetectorConfig{},
			src: fakeSource{
				query:  map[string]string{"lang": "klingon"},
				cookie: map[string]string{"lang": "ja"},
			},
			want: LangJA,
		},
		{
			name:   "an empty value falls through to the next method",
			config: DetectorConfig{},
			src: fakeSource{
				query:  map[string]string{"lang": ""},
				cookie: map[string]string{"lang": "ja"},
			},
			want: LangJA,
		},
		{
			name:   "AcceptLanguage=false skips the accept step",
			config: DetectorConfig{AcceptLanguage: false},
			src:    fakeSource{header: map[string]string{"Accept-Language": "fr"}},
			want:   DefaultLanguage,
		},
		{
			name:   "a method left out of Priority is never consulted",
			config: DetectorConfig{Priority: []string{"cookie"}},
			src: fakeSource{
				query:  map[string]string{"lang": "zh"},
				cookie: map[string]string{"lang": "ja"},
			},
			want: LangJA,
		},
		{
			name:   "an unknown Priority entry is ignored, not fatal",
			config: DetectorConfig{Priority: []string{"telepathy", "cookie"}},
			src:    fakeSource{cookie: map[string]string{"lang": "ja"}},
			want:   LangJA,
		},
		{
			name:   "custom parameter names are honoured",
			config: DetectorConfig{QueryParam: "locale", CookieName: "site_lang", HeaderName: "X-Locale"},
			src: fakeSource{
				header: map[string]string{"X-Locale": "ko"},
			},
			want: LangKO,
		},
		{
			name:   "regional variants normalise to the base language",
			config: DetectorConfig{},
			src:    fakeSource{query: map[string]string{"lang": "zh-CN"}},
			want:   LangZH,
		},
		{
			name:   "Accept-Language honours q-values, not document order",
			config: DetectorConfig{Priority: []string{"accept"}, AcceptLanguage: true},
			src:    fakeSource{header: map[string]string{"Accept-Language": "en;q=0.3,ko;q=0.9"}},
			want:   LangKO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := tt.src
			assert.Equal(t, tt.want, NewDetector(tt.config).Detect(&src))
		})
	}
}

// NewDetector fills in every zero-valued field except AcceptLanguage, which is
// a bool it cannot tell apart from "not set". So a zero DetectorConfig keeps
// "accept" in the default Priority while leaving the step switched off, and
// Accept-Language is silently ignored. Pinned because it is surprising: reach
// for DefaultDetectorConfig() (or set AcceptLanguage explicitly) when you want
// the documented defaults.
func TestNewDetector_ZeroConfigLeavesAcceptLanguageOff(t *testing.T) {
	src := &fakeSource{header: map[string]string{"Accept-Language": "fr"}}

	assert.Equal(t, DefaultLanguage, NewDetector(DetectorConfig{}).Detect(src))
	assert.Equal(t, LangFR, NewDetector(DefaultDetectorConfig()).Detect(src))
	assert.Equal(t, LangFR, DefaultDetector.Detect(src))
}

// A method missing from Priority must not even be read: an adapter's lookup can
// be expensive (or, on a framework that parses lazily, have side effects), and
// the old two-copy chain had no way to state this once.
func TestDetect_ConsultsOnlyConfiguredMethods(t *testing.T) {
	src := &fakeSource{cookie: map[string]string{"lang": "ja"}}

	lang := NewDetector(DetectorConfig{Priority: []string{"cookie"}}).Detect(src)

	assert.Equal(t, LangJA, lang)
	assert.Equal(t, []string{"lang"}, src.cookies)
	assert.Empty(t, src.queries)
	assert.Empty(t, src.headers)
}

// The accept step reads "Accept-Language", never the configured HeaderName --
// the two are separate steps that happen to share the Header lookup.
func TestDetect_AcceptStepReadsAcceptLanguageHeader(t *testing.T) {
	src := &fakeSource{}

	NewDetector(DetectorConfig{Priority: []string{"header", "accept"}, AcceptLanguage: true, HeaderName: "X-Locale"}).Detect(src)

	assert.Equal(t, []string{"X-Locale", "Accept-Language"}, src.headers)
}

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
	assert.Equal(t, LangDE, DefaultDetector.DetectFromRequest(req))
}

// DetectFromRequest is the same chain as Detect, reached through the net/http
// source rather than a hand-written one.
func TestDetectFromRequest_MatchesDetectOverStdSource(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?lang=ko", nil)

	assert.Equal(t,
		DefaultDetector.Detect(RequestSourceOf(req)),
		DefaultDetector.DetectFromRequest(req),
	)
}

// The returned RequestSource reads the request live rather than snapshotting
// it, so an adapter may hold one for the life of a request.
func TestRequestSourceOf_ReadsRequestLive(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	src := RequestSourceOf(req)

	require.Empty(t, src.Query("lang"))
	req.URL.RawQuery = url.Values{"lang": {"zh"}}.Encode()

	assert.Equal(t, "zh", src.Query("lang"))
}
