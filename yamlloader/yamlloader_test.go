package yamlloader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	i18n "github.com/soulteary/i18n-kit/v4"
)

func TestBundle_LoadYAML(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	yamlData := []byte(`
greeting: Hello
farewell: Goodbye
`)

	err := Load(bundle, i18n.LangEN, yamlData)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(i18n.LangEN, "greeting"))
	assert.Equal(t, "Goodbye", bundle.GetTranslation(i18n.LangEN, "farewell"))
}

func TestBundle_LoadYAML_Invalid(t *testing.T) {
	bundle := i18n.NewBundle(i18n.LangEN)
	yamlData := []byte(`invalid: [yaml`)

	err := Load(bundle, i18n.LangEN, yamlData)
	assert.Error(t, err)
}

func TestBundle_LoadYAMLFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "en.yaml")
	err := os.WriteFile(filePath, []byte("greeting: Hello"), 0600)
	require.NoError(t, err)

	bundle := i18n.NewBundle(i18n.LangEN)
	err = LoadFile(bundle, i18n.LangEN, filePath)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(i18n.LangEN, "greeting"))
}

func TestBundle_LoadDirectory(t *testing.T) {
	// Create temp directory with translation files
	tmpDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir, "en.json"), []byte(`{"greeting": "Hello"}`), 0600)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "zh.yaml"), []byte("greeting: 你好"), 0600)
	require.NoError(t, err)

	// Create a file that should be skipped (unknown language)
	err = os.WriteFile(filepath.Join(tmpDir, "unknown.json"), []byte(`{"greeting": "?"}`), 0600)
	require.NoError(t, err)

	// Create a non-translation file that should be skipped
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0600)
	require.NoError(t, err)

	bundle := i18n.NewBundle(i18n.LangEN)
	err = LoadDirectory(bundle, tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(i18n.LangEN, "greeting"))
	assert.Equal(t, "你好", bundle.GetTranslation(i18n.LangZH, "greeting"))
}
