package i18n

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
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

	// AcceptLanguage enables parsing of Accept-Language header.
	// Default: true
	AcceptLanguage bool

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
		QueryParam:     "lang",
		CookieName:     "lang",
		HeaderName:     "X-Language",
		AcceptLanguage: true,
		Priority:       []string{"query", "cookie", "header", "accept"},
		Default:        DefaultLanguage,
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

// DetectFromRequest detects language from a net/http request.
func (d *Detector) DetectFromRequest(r *http.Request) Language {
	for _, method := range d.config.Priority {
		var lang Language
		var found bool

		switch method {
		case "query":
			lang, found = d.detectFromQueryStd(r)
		case "cookie":
			lang, found = d.detectFromCookieStd(r)
		case "header":
			lang, found = d.detectFromHeaderStd(r)
		case "accept":
			if d.config.AcceptLanguage {
				lang, found = d.detectFromAcceptLanguageStd(r)
			}
		}

		if found && lang.IsValid() {
			return lang
		}
	}

	return d.config.Default
}

// DetectFromFiber detects language from a Fiber context.
func (d *Detector) DetectFromFiber(c *fiber.Ctx) Language {
	for _, method := range d.config.Priority {
		var lang Language
		var found bool

		switch method {
		case "query":
			lang, found = d.detectFromQueryFiber(c)
		case "cookie":
			lang, found = d.detectFromCookieFiber(c)
		case "header":
			lang, found = d.detectFromHeaderFiber(c)
		case "accept":
			if d.config.AcceptLanguage {
				lang, found = d.detectFromAcceptLanguageFiber(c)
			}
		}

		if found && lang.IsValid() {
			return lang
		}
	}

	return d.config.Default
}

// net/http detection methods

func (d *Detector) detectFromQueryStd(r *http.Request) (Language, bool) {
	if r.URL == nil {
		return "", false
	}
	value := r.URL.Query().Get(d.config.QueryParam)
	if value == "" {
		return "", false
	}
	return ParseLanguage(value)
}

func (d *Detector) detectFromCookieStd(r *http.Request) (Language, bool) {
	cookie, err := r.Cookie(d.config.CookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return ParseLanguage(cookie.Value)
}

func (d *Detector) detectFromHeaderStd(r *http.Request) (Language, bool) {
	value := r.Header.Get(d.config.HeaderName)
	if value == "" {
		return "", false
	}
	return ParseLanguage(value)
}

func (d *Detector) detectFromAcceptLanguageStd(r *http.Request) (Language, bool) {
	header := r.Header.Get("Accept-Language")
	if header == "" {
		return "", false
	}
	return parseAcceptLanguage(header)
}

// Fiber detection methods

func (d *Detector) detectFromQueryFiber(c *fiber.Ctx) (Language, bool) {
	value := c.Query(d.config.QueryParam)
	if value == "" {
		return "", false
	}
	return ParseLanguage(value)
}

func (d *Detector) detectFromCookieFiber(c *fiber.Ctx) (Language, bool) {
	value := c.Cookies(d.config.CookieName)
	if value == "" {
		return "", false
	}
	return ParseLanguage(value)
}

func (d *Detector) detectFromHeaderFiber(c *fiber.Ctx) (Language, bool) {
	value := c.Get(d.config.HeaderName)
	if value == "" {
		return "", false
	}
	return ParseLanguage(value)
}

func (d *Detector) detectFromAcceptLanguageFiber(c *fiber.Ctx) (Language, bool) {
	header := c.Get("Accept-Language")
	if header == "" {
		return "", false
	}
	return parseAcceptLanguage(header)
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

// DetectFromRequest is a convenience function using the default detector.
func DetectFromRequest(r *http.Request) Language {
	return DefaultDetector.DetectFromRequest(r)
}

// DetectFromFiber is a convenience function using the default detector.
func DetectFromFiber(c *fiber.Ctx) Language {
	return DefaultDetector.DetectFromFiber(c)
}
