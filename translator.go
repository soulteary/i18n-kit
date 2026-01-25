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

// Tf returns a formatted translated string with arguments.
// Uses the translator's current language.
func (t *Translator) Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(t.T(key), args...)
}

// TWithLang returns the translated string for a specific language.
func (t *Translator) TWithLang(lang Language, key string) string {
	return t.bundle.GetTranslation(lang, key)
}

// TfWithLang returns a formatted translated string for a specific language.
func (t *Translator) TfWithLang(lang Language, key string, args ...interface{}) string {
	return fmt.Sprintf(t.TWithLang(lang, key), args...)
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

// Tf returns a formatted translated string using the global translator.
func Tf(key string, args ...interface{}) string {
	return fmt.Sprintf(T(key), args...)
}

// TWithLang returns the translated string for a specific language.
func TWithLang(lang Language, key string) string {
	return DefaultBundle.GetTranslation(lang, key)
}

// TfWithLang returns a formatted translated string for a specific language.
func TfWithLang(lang Language, key string, args ...interface{}) string {
	return fmt.Sprintf(TWithLang(lang, key), args...)
}

// AddTranslation adds a translation to the default bundle.
func AddTranslation(lang Language, key, value string) {
	DefaultBundle.AddTranslation(lang, key, value)
}

// AddTranslations adds multiple translations to the default bundle.
func AddTranslations(lang Language, translations map[string]string) {
	DefaultBundle.AddTranslations(lang, translations)
}
