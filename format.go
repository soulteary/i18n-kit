package i18n

import (
	"context"
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

// FormatWithContext returns the translated string for the language carried by
// ctx, with the same named-parameter substitution as Format.
//
// The language comes from LanguageFromContext, so a context carrying none --
// including a nil one -- formats in DefaultLanguage. The bundle is always the
// formatter's own; a bundle stored with ContextWithBundle is not consulted,
// because a Formatter is built around one bundle and Format uses it
// unconditionally.
//
// The httpadapter middleware stores the detected language in the request
// context, so r.Context() is what a net/http handler passes here. Fiber keeps
// it in Locals rather than in a context: read it with fiberadapter.Language(c)
// and call Format directly.
//
// Until v4.0.0 this took an interface{} and type-switched for a Fiber context.
// Nothing satisfied that switch -- fiber.Ctx spells the method
// Context() context.Context, not Context() interface{} -- so every call,
// one passing a real context.Context included, quietly formatted in
// DefaultLanguage.
func (f *Formatter) FormatWithContext(ctx context.Context, key string, params map[string]interface{}) string {
	return f.Format(LanguageFromContext(ctx), key, params)
}

// substituteParams replaces {name} placeholders with values from the map.
//
// Substitution is a SINGLE pass over the template, so a substituted value is
// never rescanned. Replacing one parameter at a time with strings.ReplaceAll,
// as this used to, had two consequences: a value containing "{other}" got
// substituted again on a later iteration (placeholder injection), and because
// Go randomises map iteration order, whether that happened varied from run to
// run on identical input.
//
// A placeholder with no matching parameter is left in place, so a missing
// value is visible rather than silently blank.
func substituteParams(template string, params map[string]interface{}) string {
	if len(params) == 0 {
		return template
	}

	return TemplatePattern.ReplaceAllStringFunc(template, func(match string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(match, "{"), "}")
		value, ok := params[name]
		if !ok {
			return match
		}
		return fmt.Sprintf("%v", value)
	})
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
//
// The character class excludes "{" as well as "}", so a literal opening brace
// cannot be swallowed into a placeholder. With `[^}]+` the pattern applied to
//
//	{"message":"Hello {name}"}
//
// started at the JSON brace, ran through the placeholder's closing brace, and
// looked up the nonexistent parameter `"message":"Hello {name` -- leaving the
// real {name} unsubstituted.
var TemplatePattern = regexp.MustCompile(`\{([^{}]+)\}`)

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
