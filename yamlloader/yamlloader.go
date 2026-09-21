// Package yamlloader loads i18n-kit translations from YAML.
//
// It lives in its own package so that importing the root package does not pull
// a YAML library into programs whose translations are all JSON -- which is most
// of them, since JSON needs nothing but the standard library. Only importing
// this package links gopkg.in/yaml.v3.
//
// It adds no behaviour of its own: Load is Bundle.LoadJSON's counterpart,
// LoadDirectory is Bundle.LoadDirectory with two more extensions, and the
// parsing rules, the language-code matching and the error text all come from
// the root package.
package yamlloader

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	i18n "github.com/soulteary/i18n-kit/v4"
)

// Decode parses a YAML mapping of string keys to string values.
//
// It satisfies i18n.Decoder, so it can be handed to
// i18n.Bundle.LoadDirectoryWith alongside decoders for other formats.
func Decode(data []byte) (map[string]string, error) {
	var translations map[string]string
	if err := yaml.Unmarshal(data, &translations); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return translations, nil
}

// Load loads translations for one language from a YAML byte slice.
// It is the YAML counterpart of i18n.Bundle.LoadJSON.
func Load(b *i18n.Bundle, lang i18n.Language, data []byte) error {
	translations, err := Decode(data)
	if err != nil {
		return err
	}
	b.AddTranslations(lang, translations)
	return nil
}

// LoadFile loads translations for one language from a YAML file.
// It is the YAML counterpart of i18n.Bundle.LoadJSONFile.
func LoadFile(b *i18n.Bundle, lang i18n.Language, path string) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return Load(b, lang, data)
}

// LoadDirectory loads every "{language}.json", "{language}.yaml" and
// "{language}.yml" file in dir.
//
// This is i18n.Bundle.LoadDirectory plus the two YAML extensions, and is what
// a mixed or all-YAML directory wants. JSON stays included so that moving one
// file from .json to .yaml does not silently stop loading the rest.
func LoadDirectory(b *i18n.Bundle, dir string) error {
	return b.LoadDirectoryWith(dir, map[string]i18n.Decoder{
		".json": i18n.DecodeJSON,
		".yaml": Decode,
		".yml":  Decode,
	})
}
