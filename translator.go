package i18n

import (
	"fmt"
	"sync"
)

// Translator provides translation functionality.
// It wraps a Bundle and adds convenience methods.
type Translator struct {
	bundle *Bundle
	mu     sync.RWMutex
	lang   Language
}

// NewTranslator creates a new Translator with the specified bundle.
func NewTranslator(bundle *Bundle) *Translator {
	if bundle == nil {
		bundle = DefaultBundle
	}
	return &Translator{
		bundle: bundle,
		lang:   bundle.GetFallback(),
	}
}

// NewTranslatorWithLanguage creates a Translator with a specific language.
func NewTranslatorWithLanguage(bundle *Bundle, lang Language) *Translator {
	t := NewTranslator(bundle)
	t.lang = lang
	return t
}

// SetLanguage sets the current language for this translator.
func (t *Translator) SetLanguage(lang Language) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lang = lang
}

// GetLanguage returns the current language.
func (t *Translator) GetLanguage() Language {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lang
}

// T returns the translated string for the given key.
// Uses the translator's current language.
func (t *Translator) T(key string) string {
	t.mu.RLock()
	lang := t.lang
	t.mu.RUnlock()
	return t.bundle.GetTranslation(lang, key)
}

// formatTranslation applies args to a translation.
//
// A translation is used as a printf format string, and GetTranslation returns
// the KEY when no translation exists. Handing that key to fmt.Sprintf with
// arguments appends "%!(EXTRA string=...)": the arguments -- typically an
// email address, phone number or user id -- end up inside the message shown to
// whoever triggered it. A single missing translation therefore became an
// information leak. When the key is missing, the key alone is returned.
func formatTranslation(text string, found bool, args ...interface{}) string {
	if !found {
		return text
	}
	// A present translation is still formatted with no arguments, because it
	// may carry printf escapes of its own: "Save 10%%" has always rendered as
	// "Save 10%".
	return fmt.Sprintf(text, args...)
}

// Tf returns a formatted translated string with arguments.
// Uses the translator's current language.
//
// If the key has no translation, the key is returned as-is and the arguments
// are NOT applied; see formatTranslation.
func (t *Translator) Tf(key string, args ...interface{}) string {
	t.mu.RLock()
	lang := t.lang
	t.mu.RUnlock()

	text, found := t.bundle.LookupTranslation(lang, key)
	return formatTranslation(text, found, args...)
}

// TWithLang returns the translated string for a specific language.
func (t *Translator) TWithLang(lang Language, key string) string {
	return t.bundle.GetTranslation(lang, key)
}

// TfWithLang returns a formatted translated string for a specific language.
//
// If the key has no translation, the key is returned as-is and the arguments
// are NOT applied; see formatTranslation.
func (t *Translator) TfWithLang(lang Language, key string, args ...interface{}) string {
	text, found := t.bundle.LookupTranslation(lang, key)
	return formatTranslation(text, found, args...)
}

// Bundle returns the underlying translation bundle.
func (t *Translator) Bundle() *Bundle {
	return t.bundle
}

// GlobalTranslator is a convenience global translator.
var GlobalTranslator = NewTranslator(DefaultBundle)

// globalLang is the global language setting (thread-safe).
var (
	globalLang   = LangEN
	globalLangMu sync.RWMutex
)

// SetGlobalLanguage sets the global language.
func SetGlobalLanguage(lang Language) {
	globalLangMu.Lock()
	defer globalLangMu.Unlock()
	if lang.IsValid() {
		globalLang = lang
	}
}

// GetGlobalLanguage returns the current global language.
func GetGlobalLanguage() Language {
	globalLangMu.RLock()
	defer globalLangMu.RUnlock()
	return globalLang
}

// T returns the translated string using the global translator and language.
func T(key string) string {
	globalLangMu.RLock()
	lang := globalLang
	globalLangMu.RUnlock()
	return DefaultBundle.GetTranslation(lang, key)
}

// Tf returns a formatted translated string using the global language.
//
// This reads globalLang, the same setting T reads and SetGlobalLanguage
// writes. It used to read GlobalTranslator's own language, which
// SetGlobalLanguage never touches, so after SetGlobalLanguage(LangZH) T
// returned Chinese while Tf returned the English fallback.
func Tf(key string, args ...interface{}) string {
	globalLangMu.RLock()
	lang := globalLang
	globalLangMu.RUnlock()
	text, found := DefaultBundle.LookupTranslation(lang, key)
	return formatTranslation(text, found, args...)
}

// TWithLang returns the translated string for a specific language.
func TWithLang(lang Language, key string) string {
	return DefaultBundle.GetTranslation(lang, key)
}

// TfWithLang returns a formatted translated string for a specific language.
func TfWithLang(lang Language, key string, args ...interface{}) string {
	text, found := DefaultBundle.LookupTranslation(lang, key)
	return formatTranslation(text, found, args...)
}

// AddTranslation adds a translation to the default bundle.
func AddTranslation(lang Language, key, value string) {
	DefaultBundle.AddTranslation(lang, key, value)
}

// AddTranslations adds multiple translations to the default bundle.
func AddTranslations(lang Language, translations map[string]string) {
	DefaultBundle.AddTranslations(lang, translations)
}
