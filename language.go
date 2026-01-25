// Package i18n provides internationalization (i18n) support for Go applications.
// It supports multiple languages, language detection from HTTP requests, and
// both Fiber and net/http middleware.
//
//nolint:misspell // This file contains multi-language translations
package i18n

import (
	"strings"
)

// Language represents a supported language code.
type Language string

// Supported language codes.
const (
	// LangEN is English (default)
	LangEN Language = "en"
	// LangZH is Chinese
	LangZH Language = "zh"
	// LangFR is French
	LangFR Language = "fr"
	// LangIT is Italian
	LangIT Language = "it"
	// LangJA is Japanese
	LangJA Language = "ja"
	// LangDE is German
	LangDE Language = "de"
	// LangKO is Korean
	LangKO Language = "ko"
	// LangES is Spanish
	LangES Language = "es"
	// LangPT is Portuguese
	LangPT Language = "pt"
	// LangRU is Russian
	LangRU Language = "ru"
)

// DefaultLanguage is the fallback language when detection fails.
var DefaultLanguage = LangEN

// SupportedLanguages contains all built-in supported languages.
var SupportedLanguages = []Language{
	LangEN, LangZH, LangFR, LangIT, LangJA, LangDE, LangKO, LangES, LangPT, LangRU,
}

// String returns the string representation of the language.
func (l Language) String() string {
	return string(l)
}

// IsValid checks if the language is in the supported languages list.
func (l Language) IsValid() bool {
	for _, lang := range SupportedLanguages {
		if lang == l {
			return true
		}
	}
	return false
}

// languageAliases maps common language codes and locales to their base language.
var languageAliases = map[string]Language{
	// English
	"en":    LangEN,
	"en-us": LangEN,
	"en_us": LangEN,
	"en-gb": LangEN,
	"en_gb": LangEN,
	"en-au": LangEN,
	"en_au": LangEN,
	// Chinese
	"zh":      LangZH,
	"zh-cn":   LangZH,
	"zh_cn":   LangZH,
	"zh-hans": LangZH,
	"zh-tw":   LangZH,
	"zh_tw":   LangZH,
	"zh-hant": LangZH,
	// French
	"fr":    LangFR,
	"fr-fr": LangFR,
	"fr_fr": LangFR,
	"fr-ca": LangFR,
	"fr_ca": LangFR,
	// Italian
	"it":    LangIT,
	"it-it": LangIT,
	"it_it": LangIT,
	// Japanese
	"ja":    LangJA,
	"ja-jp": LangJA,
	"ja_jp": LangJA,
	// German
	"de":    LangDE,
	"de-de": LangDE,
	"de_de": LangDE,
	"de-at": LangDE,
	"de_at": LangDE,
	"de-ch": LangDE,
	"de_ch": LangDE,
	// Korean
	"ko":    LangKO,
	"ko-kr": LangKO,
	"ko_kr": LangKO,
	// Spanish
	"es":    LangES,
	"es-es": LangES,
	"es_es": LangES,
	"es-mx": LangES,
	"es_mx": LangES,
	// Portuguese
	"pt":    LangPT,
	"pt-pt": LangPT,
	"pt_pt": LangPT,
	"pt-br": LangPT,
	"pt_br": LangPT,
	// Russian
	"ru":    LangRU,
	"ru-ru": LangRU,
	"ru_ru": LangRU,
}

// NormalizeLanguage normalizes a language code to a supported Language.
// It handles common variants like "zh-CN", "en-US", "fr_FR", etc.
// Returns DefaultLanguage if the code is not recognized.
func NormalizeLanguage(code string) Language {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return DefaultLanguage
	}

	// Try direct lookup
	if lang, ok := languageAliases[code]; ok {
		return lang
	}

	// Try base language code (before hyphen or underscore)
	if idx := strings.IndexAny(code, "-_"); idx > 0 {
		base := code[:idx]
		if lang, ok := languageAliases[base]; ok {
			return lang
		}
	}

	return DefaultLanguage
}

// ParseLanguage parses a language code string to a Language type.
// Unlike NormalizeLanguage, it returns (Language, false) if not recognized.
func ParseLanguage(code string) (Language, bool) {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return "", false
	}

	if lang, ok := languageAliases[code]; ok {
		return lang, true
	}

	if idx := strings.IndexAny(code, "-_"); idx > 0 {
		base := code[:idx]
		if lang, ok := languageAliases[base]; ok {
			return lang, true
		}
	}

	return "", false
}

// AddLanguageAlias adds a custom alias for a language.
// This allows applications to register their own locale variants.
func AddLanguageAlias(alias string, lang Language) {
	languageAliases[strings.ToLower(alias)] = lang
}

// RegisterLanguage registers a new supported language.
func RegisterLanguage(lang Language) {
	for _, existing := range SupportedLanguages {
		if existing == lang {
			return
		}
	}
	SupportedLanguages = append(SupportedLanguages, lang)
}
