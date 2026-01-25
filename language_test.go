package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLanguage_String(t *testing.T) {
	tests := []struct {
		lang     Language
		expected string
	}{
		{LangEN, "en"},
		{LangZH, "zh"},
		{LangFR, "fr"},
		{LangDE, "de"},
		{LangJA, "ja"},
		{LangKO, "ko"},
		{LangIT, "it"},
		{LangES, "es"},
		{LangPT, "pt"},
		{LangRU, "ru"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.lang.String())
		})
	}
}

func TestLanguage_IsValid(t *testing.T) {
	tests := []struct {
		lang    Language
		isValid bool
	}{
		{LangEN, true},
		{LangZH, true},
		{LangFR, true},
		{Language("xx"), false},
		{Language(""), false},
		{Language("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.lang.IsValid())
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected Language
	}{
		// English variants
		{"en", LangEN},
		{"en-US", LangEN},
		{"en-GB", LangEN},
		{"en_us", LangEN},
		{"EN", LangEN},
		// Chinese variants
		{"zh", LangZH},
		{"zh-CN", LangZH},
		{"zh-TW", LangZH},
		{"zh_cn", LangZH},
		{"zh-Hans", LangZH},
		// French variants
		{"fr", LangFR},
		{"fr-FR", LangFR},
		{"fr-CA", LangFR},
		// German variants
		{"de", LangDE},
		{"de-DE", LangDE},
		{"de-AT", LangDE},
		// Japanese
		{"ja", LangJA},
		{"ja-JP", LangJA},
		// Korean
		{"ko", LangKO},
		{"ko-KR", LangKO},
		// Italian
		{"it", LangIT},
		{"it-IT", LangIT},
		// Spanish
		{"es", LangES},
		{"es-ES", LangES},
		{"es-MX", LangES},
		// Portuguese
		{"pt", LangPT},
		{"pt-BR", LangPT},
		// Russian
		{"ru", LangRU},
		{"ru-RU", LangRU},
		// Edge cases
		{"", LangEN},
		{"  ", LangEN},
		{"unknown", LangEN},
		{"xx-YY", LangEN},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected Language
		ok       bool
	}{
		{"en", LangEN, true},
		{"zh-CN", LangZH, true},
		{"fr-FR", LangFR, true},
		{"", "", false},
		{"unknown", "", false},
		{"xx-YY", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lang, ok := ParseLanguage(tt.input)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, lang)
			}
		})
	}
}

func TestAddLanguageAlias(t *testing.T) {
	// Add custom alias
	AddLanguageAlias("en-au", LangEN)
	AddLanguageAlias("custom", LangZH)

	tests := []struct {
		input    string
		expected Language
	}{
		{"en-au", LangEN},
		{"custom", LangZH},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegisterLanguage(t *testing.T) {
	// Register new language
	newLang := Language("ar")
	RegisterLanguage(newLang)

	// Check it's in the list
	found := false
	for _, lang := range SupportedLanguages {
		if lang == newLang {
			found = true
			break
		}
	}
	assert.True(t, found, "new language should be registered")

	// Registering again should not duplicate
	count := 0
	for _, lang := range SupportedLanguages {
		if lang == newLang {
			count++
		}
	}
	assert.Equal(t, 1, count, "language should not be duplicated")

	// Register again - should not add duplicate
	RegisterLanguage(newLang)
	count = 0
	for _, lang := range SupportedLanguages {
		if lang == newLang {
			count++
		}
	}
	assert.Equal(t, 1, count, "duplicate registration should be ignored")
}
