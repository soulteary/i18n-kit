package i18n

import (
	"fmt"
	"regexp"
	"strings"
)

// Formatter provides advanced formatting capabilities for translations.
type Formatter struct {
	bundle *Bundle
}

// NewFormatter creates a new Formatter with the given bundle.
func NewFormatter(bundle *Bundle) *Formatter {
	if bundle == nil {
		bundle = DefaultBundle
	}
	return &Formatter{bundle: bundle}
}

// DefaultFormatter is a formatter using the default bundle.
var DefaultFormatter = NewFormatter(DefaultBundle)

// Format returns the translated string with named parameter substitution.
// Parameters are specified as {name} in the translation string.
// Example: Format(LangEN, "greeting", map[string]interface{}{"name": "World"})
// With translation "Hello, {name}!" -> "Hello, World!"
func (f *Formatter) Format(lang Language, key string, params map[string]interface{}) string {
	translation := f.bundle.GetTranslation(lang, key)
	return substituteParams(translation, params)
}

// FormatWithContext returns the translated string using language from context.
func (f *Formatter) FormatWithContext(ctx interface{}, key string, params map[string]interface{}) string {
	// Type assertion for different context types
	switch c := ctx.(type) {
	case interface{ Context() interface{} }:
		// Fiber context - check locals
		if lang, ok := c.(interface {
			Locals(key interface{}) interface{}
		}); ok {
			if l, ok := lang.Locals("i18n-language").(Language); ok {
				return f.Format(l, key, params)
			}
		}
	}
	return f.Format(DefaultLanguage, key, params)
}

// substituteParams replaces {name} placeholders with values from the map.
func substituteParams(template string, params map[string]interface{}) string {
	if len(params) == 0 {
		return template
	}

	result := template
	for name, value := range params {
		placeholder := "{" + name + "}"
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// PluralizationRule defines a rule for plural forms.
type PluralizationRule struct {
	// Zero is used when count is 0 (optional)
	Zero string
	// One is used when count is 1
	One string
	// Few is used for few items (language-specific, optional)
	Few string
	// Many is used for many items (language-specific, optional)
	Many string
	// Other is the default plural form
	Other string
}

// Pluralize returns the appropriate plural form based on count.
// This is a simple implementation that works for English-like languages.
func (f *Formatter) Pluralize(lang Language, key string, count int, params map[string]interface{}) string {
	// Get translations for different plural forms
	zeroKey := key + ".zero"
	oneKey := key + ".one"
	otherKey := key + ".other"

	var translation string

	switch {
	case count == 0 && f.bundle.HasTranslation(lang, zeroKey):
		translation = f.bundle.GetTranslation(lang, zeroKey)
	case count == 1 && f.bundle.HasTranslation(lang, oneKey):
		translation = f.bundle.GetTranslation(lang, oneKey)
	default:
		translation = f.bundle.GetTranslation(lang, otherKey)
		// Fallback to base key if .other doesn't exist
		if translation == otherKey {
			translation = f.bundle.GetTranslation(lang, key)
		}
	}

	// Add count to params
	if params == nil {
		params = make(map[string]interface{})
	}
	params["count"] = count

	return substituteParams(translation, params)
}

// PluralizeSimple returns the appropriate plural form for simple cases.
// singular is used for count=1, plural for all other values.
func PluralizeSimple(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// Format is a convenience function using the default formatter.
func Format(lang Language, key string, params map[string]interface{}) string {
	return DefaultFormatter.Format(lang, key, params)
}

// Pluralize is a convenience function using the default formatter.
func Pluralize(lang Language, key string, count int, params map[string]interface{}) string {
	return DefaultFormatter.Pluralize(lang, key, count, params)
}

// TemplatePattern matches {variable} patterns.
var TemplatePattern = regexp.MustCompile(`\{([^}]+)\}`)

// ExtractParams extracts parameter names from a translation template.
func ExtractParams(template string) []string {
	matches := TemplatePattern.FindAllStringSubmatch(template, -1)
	params := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			params = append(params, match[1])
			seen[match[1]] = true
		}
	}
	return params
}

// HasParams checks if a translation template contains parameters.
func HasParams(template string) bool {
	return TemplatePattern.MatchString(template)
}

// ValidateParams checks if all required parameters are provided.
func ValidateParams(template string, params map[string]interface{}) (missing []string) {
	required := ExtractParams(template)
	for _, name := range required {
		if _, ok := params[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing
}
