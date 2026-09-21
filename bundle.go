package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Bundle manages translations for multiple languages.
// It is safe for concurrent use.
type Bundle struct {
	mu           sync.RWMutex
	translations map[Language]map[string]string
	fallback     Language
}

// NewBundle creates a new translation bundle with the specified fallback language.
func NewBundle(fallback Language) *Bundle {
	return &Bundle{
		translations: make(map[Language]map[string]string),
		fallback:     fallback,
	}
}

// DefaultBundle is a global bundle for convenience.
var DefaultBundle = NewBundle(LangEN)

// SetFallback sets the fallback language for the bundle.
func (b *Bundle) SetFallback(lang Language) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.fallback = lang
}

// GetFallback returns the current fallback language.
func (b *Bundle) GetFallback() Language {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.fallback
}

// AddTranslation adds a single translation to the bundle.
func (b *Bundle) AddTranslation(lang Language, key, value string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.translations[lang] == nil {
		b.translations[lang] = make(map[string]string)
	}
	b.translations[lang][key] = value
}

// AddTranslations adds multiple translations for a language.
func (b *Bundle) AddTranslations(lang Language, translations map[string]string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.translations[lang] == nil {
		b.translations[lang] = make(map[string]string)
	}
	for k, v := range translations {
		b.translations[lang][k] = v
	}
}

// GetTranslation retrieves a translation for the given language and key.
// It falls back to the fallback language if not found.
// Returns the key itself if no translation exists.
// LookupTranslation is GetTranslation with an explicit found flag.
//
// GetTranslation returns the key itself when nothing is found, which is a fine
// fallback for display but indistinguishable from a translation that happens
// to equal its key -- and callers that then use the result as a format string
// need to know the difference. See Translator.Tf.
func (b *Bundle) LookupTranslation(lang Language, key string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Try the requested language
	if langMap, ok := b.translations[lang]; ok {
		if translation, ok := langMap[key]; ok {
			return translation, true
		}
	}

	// Try fallback language
	if lang != b.fallback {
		if langMap, ok := b.translations[b.fallback]; ok {
			if translation, ok := langMap[key]; ok {
				return translation, true
			}
		}
	}

	return key, false
}

// GetTranslation returns the translation of key in lang, falling back to the
// bundle's fallback language and then to key itself when neither has it.
//
// Because a missing key comes back as the key, the result is indistinguishable
// from a translation that happens to equal its key. Use [Bundle.LookupTranslation]
// when you need to tell those apart -- which is what the Tf family does, so
// that a missing key is never used as a printf format string.
func (b *Bundle) GetTranslation(lang Language, key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Try the requested language
	if langMap, ok := b.translations[lang]; ok {
		if translation, ok := langMap[key]; ok {
			return translation
		}
	}

	// Try fallback language
	if lang != b.fallback {
		if langMap, ok := b.translations[b.fallback]; ok {
			if translation, ok := langMap[key]; ok {
				return translation
			}
		}
	}

	// Return key if not found
	return key
}

// HasTranslation checks if a translation exists for the given language and key.
func (b *Bundle) HasTranslation(lang Language, key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if langMap, ok := b.translations[lang]; ok {
		_, exists := langMap[key]
		return exists
	}
	return false
}

// Languages returns all languages that have translations in the bundle.
func (b *Bundle) Languages() []Language {
	b.mu.RLock()
	defer b.mu.RUnlock()

	langs := make([]Language, 0, len(b.translations))
	for lang := range b.translations {
		langs = append(langs, lang)
	}
	return langs
}

// Keys returns all translation keys for a given language.
func (b *Bundle) Keys(lang Language) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	langMap, ok := b.translations[lang]
	if !ok {
		return nil
	}

	keys := make([]string, 0, len(langMap))
	for k := range langMap {
		keys = append(keys, k)
	}
	return keys
}

// Clear removes all translations from the bundle.
func (b *Bundle) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.translations = make(map[Language]map[string]string)
}

// ClearLanguage removes all translations for a specific language.
func (b *Bundle) ClearLanguage(lang Language) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.translations, lang)
}

// LoadJSON loads translations from a JSON byte slice.
// The JSON should be an object with string keys and string values.
func (b *Bundle) LoadJSON(lang Language, data []byte) error {
	translations, err := DecodeJSON(data)
	if err != nil {
		return err
	}
	b.AddTranslations(lang, translations)
	return nil
}

// LoadJSONFile loads translations from a JSON file.
func (b *Bundle) LoadJSONFile(lang Language, path string) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return b.LoadJSON(lang, data)
}

// Decoder parses one translation file's bytes into a flat key/value map.
//
// It exists so that LoadDirectoryWith can walk a directory without this package
// knowing every format: the yamlloader subpackage supplies a YAML Decoder, and
// nothing here has to import a YAML library to make that work. Anyone wanting
// TOML or .properties writes twenty lines and needs no change here either.
type Decoder func(data []byte) (map[string]string, error)

// DecodeJSON is the Decoder LoadDirectory uses for ".json" files.
func DecodeJSON(data []byte) (map[string]string, error) {
	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return translations, nil
}

// LoadDirectory loads all JSON translation files from a directory.
// Files should be named as "{language}.json" -- for example "en.json",
// "fr-FR.json".
//
// YAML is not built in: it would mean importing a YAML library into every
// program that imports this package, including the ones whose translations are
// all JSON. A directory containing .yaml or .yml files is reported as an error
// naming the fix rather than silently loaded at half strength -- use
// yamlloader.LoadDirectory, which understands all three extensions.
func (b *Bundle) LoadDirectory(dir string) error {
	return b.LoadDirectoryWith(dir, map[string]Decoder{".json": DecodeJSON})
}

// LoadDirectoryWith is LoadDirectory with the set of extensions it understands
// given explicitly, keyed by lowercase extension including the dot.
//
// Files whose extension is absent from decoders are skipped, as are files whose
// name is not a language this package knows -- with one exception: .yaml and
// .yml are reported as an error when no decoder covers them, because a
// directory of YAML translations loading as an empty bundle is the kind of
// quiet failure that reaches production.
func (b *Bundle) LoadDirectoryWith(dir string, decoders map[string]Decoder) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		decode, ok := decoders[ext]
		if !ok {
			if ext == ".yaml" || ext == ".yml" {
				return fmt.Errorf("%s: YAML translations need the yamlloader subpackage"+
					" -- call yamlloader.LoadDirectory(bundle, dir) instead", name)
			}
			continue
		}

		langCode := strings.TrimSuffix(name, ext)
		lang, ok := ParseLanguage(langCode)
		if !ok {
			// Unknown language, skip
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("failed to load %s: failed to read file: %w", name, err)
		}
		translations, err := decode(data)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", name, err)
		}
		b.AddTranslations(lang, translations)
	}

	return nil
}

// Merge merges another bundle into this one.
// Existing translations are overwritten.
func (b *Bundle) Merge(other *Bundle) {
	other.mu.RLock()
	defer other.mu.RUnlock()

	for lang, translations := range other.translations {
		b.AddTranslations(lang, translations)
	}
}

// Clone creates a deep copy of the bundle.
func (b *Bundle) Clone() *Bundle {
	b.mu.RLock()
	defer b.mu.RUnlock()

	clone := NewBundle(b.fallback)
	for lang, translations := range b.translations {
		clone.translations[lang] = make(map[string]string, len(translations))
		for k, v := range translations {
			clone.translations[lang][k] = v
		}
	}
	return clone
}
