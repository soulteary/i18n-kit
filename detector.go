package i18n

import (
	"sort"
	"strconv"
	"strings"
)

// DetectorConfig configures language detection behavior.
type DetectorConfig struct {
	// QueryParam is the query parameter name to check for language.
	// Default: "lang"
	QueryParam string

	// CookieName is the cookie name to check for language.
	// Default: "lang"
	CookieName string

	// HeaderName is the header name for explicit language override.
	// Default: "X-Language"
	HeaderName string

	// DisableAcceptLanguage turns off parsing of the Accept-Language header.
	//
	// Negative so that the zero value is the documented default. As a plain
	// AcceptLanguage bool it could not be told apart from "not set", so
	// NewDetector could not fill it in the way it fills every other field, and
	// NewDetector(DetectorConfig{}) kept "accept" in the default Priority while
	// silently leaving the step switched off.
	//
	// Default: false, i.e. Accept-Language is parsed.
	DisableAcceptLanguage bool

	// Priority defines the order of detection methods.
	// Available methods: "query", "cookie", "header", "accept"
	// Default: ["query", "cookie", "header", "accept"]
	Priority []string

	// Default is the fallback language if detection fails.
	// Default: DefaultLanguage
	Default Language
}

// DefaultDetectorConfig returns the default detector configuration.
func DefaultDetectorConfig() DetectorConfig {
	return DetectorConfig{
		QueryParam: "lang",
		CookieName: "lang",
		HeaderName: "X-Language",
		Priority:   []string{"query", "cookie", "header", "accept"},
		Default:    DefaultLanguage,
	}
}

// Detector detects language from HTTP requests.
type Detector struct {
	config DetectorConfig
}

// NewDetector creates a new language detector with the given config.
func NewDetector(config DetectorConfig) *Detector {
	if config.QueryParam == "" {
		config.QueryParam = "lang"
	}
	if config.CookieName == "" {
		config.CookieName = "lang"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-Language"
	}
	if len(config.Priority) == 0 {
		config.Priority = []string{"query", "cookie", "header", "accept"}
	}
	if config.Default == "" {
		config.Default = DefaultLanguage
	}
	return &Detector{config: config}
}

// DefaultDetector is a detector with default configuration.
var DefaultDetector = NewDetector(DefaultDetectorConfig())

// RequestSource is the minimal view of an incoming request that detection
// needs: three string lookups. Implementing it is all a framework adapter has
// to do -- see the fiberadapter subpackage.
//
// It exists so the priority chain below has exactly one implementation. It used
// to have two, one per framework, kept in step by hand; a method added to one
// and forgotten in the other would have shown up as "detection works on
// net/http but not on Fiber" and nothing else.
type RequestSource interface {
	// Query returns a URL query parameter, or "" when absent.
	Query(name string) string
	// Cookie returns a request cookie's value, or "" when absent.
	Cookie(name string) string
	// Header returns a request header, or "" when absent.
	Header(name string) string
}

// Detect runs the configured priority chain against src and returns the first
// valid language it finds, or the configured default.
func (d *Detector) Detect(src RequestSource) Language {
	for _, method := range d.config.Priority {
		var lang Language
		var found bool

		switch method {
		case "query":
			lang, found = parseDetected(src.Query(d.config.QueryParam))
		case "cookie":
			lang, found = parseDetected(src.Cookie(d.config.CookieName))
		case "header":
			lang, found = parseDetected(src.Header(d.config.HeaderName))
		case "accept":
			if !d.config.DisableAcceptLanguage {
				if header := src.Header("Accept-Language"); header != "" {
					lang, found = parseAcceptLanguage(header)
				}
			}
		}

		if found && lang.IsValid() {
			return lang
		}
	}

	return d.config.Default
}

// parseDetected treats an absent value as "not found" rather than feeding an
// empty string to ParseLanguage.
func parseDetected(value string) (Language, bool) {
	if value == "" {
		return "", false
	}
	return ParseLanguage(value)
}

// langWithQuality represents a language with its quality value.
type langWithQuality struct {
	lang    string
	quality float64
}

// parseAcceptLanguage parses the Accept-Language header and returns the best matching language.
// Format: "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7"
func parseAcceptLanguage(header string) (Language, bool) {
	if header == "" {
		return "", false
	}

	var langs []langWithQuality

	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for quality value
		subparts := strings.Split(part, ";")
		lang := strings.TrimSpace(subparts[0])
		quality := 1.0

		if len(subparts) > 1 {
			for _, sp := range subparts[1:] {
				sp = strings.TrimSpace(sp)
				if strings.HasPrefix(sp, "q=") {
					if q, err := strconv.ParseFloat(sp[2:], 64); err == nil {
						quality = q
					}
				}
			}
		}

		if lang != "" && lang != "*" {
			langs = append(langs, langWithQuality{lang: lang, quality: quality})
		}
	}

	// Sort by quality (highest first)
	sort.Slice(langs, func(i, j int) bool {
		return langs[i].quality > langs[j].quality
	})

	// Return first matching language
	for _, lq := range langs {
		if lang, ok := ParseLanguage(lq.lang); ok {
			return lang, true
		}
	}

	return "", false
}
