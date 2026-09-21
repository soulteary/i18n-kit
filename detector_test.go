package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDetectorConfig(t *testing.T) {
	config := DefaultDetectorConfig()

	assert.Equal(t, "lang", config.QueryParam)
	assert.Equal(t, "lang", config.CookieName)
	assert.Equal(t, "X-Language", config.HeaderName)
	assert.False(t, config.DisableAcceptLanguage)
	assert.Equal(t, []string{"query", "cookie", "header", "accept"}, config.Priority)
	assert.Equal(t, DefaultLanguage, config.Default)
}

func TestNewDetector(t *testing.T) {
	detector := NewDetector(DetectorConfig{})

	assert.NotNil(t, detector)
	assert.Equal(t, "lang", detector.config.QueryParam)
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
