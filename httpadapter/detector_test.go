package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	i18n "github.com/soulteary/i18n-kit/v4"
)

func TestDetector_DetectFromRequest_Query(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"query"},
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangZH, lang)
}

func TestDetector_DetectFromRequest_Cookie(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"cookie"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangFR, lang)
}

func TestDetector_DetectFromRequest_Header(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"header"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Language", "de")
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangDE, lang)
}

func TestDetector_DetectFromRequest_AcceptLanguage(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"accept"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ja-JP,ja;q=0.9,en;q=0.8")
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangJA, lang)
}

func TestDetector_DetectFromRequest_Priority(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"query", "cookie", "header", "accept"},
	})

	// Query takes priority
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "ja")
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangZH, lang)

	// Cookie when no query
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "ja")
	lang = DetectWith(detector, req)
	assert.Equal(t, i18n.LangFR, lang)

	// Header when no cookie
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "ja")
	lang = DetectWith(detector, req)
	assert.Equal(t, i18n.LangDE, lang)

	// Accept-i18n.Language when no header
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ja")
	lang = DetectWith(detector, req)
	assert.Equal(t, i18n.LangJA, lang)
}

func TestDetector_DetectFromRequest_Fallback(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"query"},
		Default:  i18n.LangFR,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangFR, lang)
}

func TestDetector_DetectFromRequest_InvalidLanguage(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"query"},
		Default:  i18n.LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=invalid", nil)
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangEN, lang)
}

func TestDetectFromRequest_Convenience(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	lang := Detect(req)
	assert.Equal(t, i18n.LangZH, lang)
}

func TestDetector_AcceptLanguageDisabled(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority:              []string{"accept"},
		DisableAcceptLanguage: true,
		Default:               i18n.LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh-CN")
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangEN, lang) // Should not detect from Accept-i18n.Language
}

func TestDetector_CustomConfig(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		QueryParam: "language",
		CookieName: "locale",
		HeaderName: "X-Locale",
		Priority:   []string{"query"},
	})

	// Should use custom query param
	req := httptest.NewRequest(http.MethodGet, "/?language=fr", nil)
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangFR, lang)

	// Default query param should not work
	req = httptest.NewRequest(http.MethodGet, "/?lang=fr", nil)
	lang = DetectWith(detector, req)
	assert.Equal(t, i18n.DefaultLanguage, lang)
}

func TestDetector_NilURL(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"query"},
		Default:  i18n.LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL = nil
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangEN, lang)
}

func TestDetector_EmptyCookie(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"cookie"},
		Default:  i18n.LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: ""})
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangEN, lang)
}

func TestDetector_EmptyHeader(t *testing.T) {
	detector := i18n.NewDetector(i18n.DetectorConfig{
		Priority: []string{"header"},
		Default:  i18n.LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Language", "")
	lang := DetectWith(detector, req)
	assert.Equal(t, i18n.LangEN, lang)
}
