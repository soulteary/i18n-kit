package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
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
	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}
	b.AddTranslations(lang, translations)
	return nil
}

// LoadYAML loads translations from a YAML byte slice.
// The YAML should be a mapping of string keys to string values.
func (b *Bundle) LoadYAML(lang Language, data []byte) error {
	var translations map[string]string
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
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

// LoadYAMLFile loads translations from a YAML file.
func (b *Bundle) LoadYAMLFile(lang Language, path string) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return b.LoadYAML(lang, data)
}

// LoadDirectory loads all translation files from a directory.
// Files should be named as "{language}.json" or "{language}.yaml".
// For example: "en.json", "zh.yaml", "fr-FR.json", etc.
func (b *Bundle) LoadDirectory(dir string) error {
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
		if ext != ".json" && ext != ".yaml" && ext != ".yml" {
			continue
		}

		langCode := strings.TrimSuffix(name, ext)
		lang, ok := ParseLanguage(langCode)
		if !ok {
			// Unknown language, skip
			continue
		}

		path := filepath.Join(dir, name)
		switch ext {
		case ".json":
			if err := b.LoadJSONFile(lang, path); err != nil {
				return fmt.Errorf("failed to load %s: %w", name, err)
			}
		case ".yaml", ".yml":
			if err := b.LoadYAMLFile(lang, path); err != nil {
				return fmt.Errorf("failed to load %s: %w", name, err)
			}
		}
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
