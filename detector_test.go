package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDetectorConfig(t *testing.T) {
	config := DefaultDetectorConfig()

	assert.Equal(t, "lang", config.QueryParam)
	assert.Equal(t, "lang", config.CookieName)
	assert.Equal(t, "X-Language", config.HeaderName)
	assert.True(t, config.AcceptLanguage)
	assert.Equal(t, []string{"query", "cookie", "header", "accept"}, config.Priority)
	assert.Equal(t, DefaultLanguage, config.Default)
}

func TestNewDetector(t *testing.T) {
	detector := NewDetector(DetectorConfig{})

	assert.NotNil(t, detector)
	assert.Equal(t, "lang", detector.config.QueryParam)
}

func TestDetector_DetectFromRequest_Query(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"query"},
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangZH, lang)
}

func TestDetector_DetectFromRequest_Cookie(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"cookie"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangFR, lang)
}

func TestDetector_DetectFromRequest_Header(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"header"},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Language", "de")
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangDE, lang)
}

func TestDetector_DetectFromRequest_AcceptLanguage(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority:       []string{"accept"},
		AcceptLanguage: true,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ja-JP,ja;q=0.9,en;q=0.8")
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangJA, lang)
}

func TestDetector_DetectFromRequest_Priority(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority:       []string{"query", "cookie", "header", "accept"},
		AcceptLanguage: true,
	})

	// Query takes priority
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "ja")
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangZH, lang)

	// Cookie when no query
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: "fr"})
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "ja")
	lang = detector.DetectFromRequest(req)
	assert.Equal(t, LangFR, lang)

	// Header when no cookie
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Language", "de")
	req.Header.Set("Accept-Language", "ja")
	lang = detector.DetectFromRequest(req)
	assert.Equal(t, LangDE, lang)

	// Accept-Language when no header
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "ja")
	lang = detector.DetectFromRequest(req)
	assert.Equal(t, LangJA, lang)
}

func TestDetector_DetectFromRequest_Fallback(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"query"},
		Default:  LangFR,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangFR, lang)
}

func TestDetector_DetectFromRequest_InvalidLanguage(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"query"},
		Default:  LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=invalid", nil)
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangEN, lang)
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		header   string
		expected Language
		ok       bool
	}{
		{"en-US,en;q=0.9,zh-CN;q=0.8", LangEN, true},
		{"zh-CN,zh;q=0.9,en;q=0.8", LangZH, true},
		{"fr-FR;q=0.8,en;q=0.9,de;q=0.7", LangEN, true}, // en has higher quality
		{"ja", LangJA, true},
		{"", "", false},
		{"*", "", false},
		{"invalid", "", false},
		{"en;q=0.5, zh;q=0.9", LangZH, true}, // zh has higher quality
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			lang, ok := parseAcceptLanguage(tt.header)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, lang)
			}
		})
	}
}

func TestDetectFromRequest_Convenience(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?lang=zh", nil)
	lang := DetectFromRequest(req)
	assert.Equal(t, LangZH, lang)
}

func TestDetector_AcceptLanguageDisabled(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority:       []string{"accept"},
		AcceptLanguage: false,
		Default:        LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "zh-CN")
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangEN, lang) // Should not detect from Accept-Language
}

func TestDetector_CustomConfig(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		QueryParam: "language",
		CookieName: "locale",
		HeaderName: "X-Locale",
		Priority:   []string{"query"},
	})

	// Should use custom query param
	req := httptest.NewRequest(http.MethodGet, "/?language=fr", nil)
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangFR, lang)

	// Default query param should not work
	req = httptest.NewRequest(http.MethodGet, "/?lang=fr", nil)
	lang = detector.DetectFromRequest(req)
	assert.Equal(t, DefaultLanguage, lang)
}

func TestDetector_NilURL(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"query"},
		Default:  LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL = nil
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangEN, lang)
}

func TestDetector_EmptyCookie(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"cookie"},
		Default:  LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "lang", Value: ""})
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangEN, lang)
}

func TestDetector_EmptyHeader(t *testing.T) {
	detector := NewDetector(DetectorConfig{
		Priority: []string{"header"},
		Default:  LangEN,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Language", "")
	lang := detector.DetectFromRequest(req)
	assert.Equal(t, LangEN, lang)
}
