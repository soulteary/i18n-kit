package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	assert.NotNil(t, bundle)
	assert.Equal(t, LangEN, bundle.GetFallback())
}

func TestBundle_SetFallback(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.SetFallback(LangZH)
	assert.Equal(t, LangZH, bundle.GetFallback())
}

func TestBundle_AddTranslation(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "你好", bundle.GetTranslation(LangZH, "greeting"))
}

func TestBundle_AddTranslations(t *testing.T) {
	bundle := NewBundle(LangEN)
	translations := map[string]string{
		"greeting": "Hello",
		"farewell": "Goodbye",
	}
	bundle.AddTranslations(LangEN, translations)

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "Goodbye", bundle.GetTranslation(LangEN, "farewell"))
}

func TestBundle_GetTranslation_Fallback(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")

	// Should fallback to English for missing Chinese translation
	result := bundle.GetTranslation(LangZH, "greeting")
	assert.Equal(t, "Hello", result)
}

func TestBundle_GetTranslation_NotFound(t *testing.T) {
	bundle := NewBundle(LangEN)

	// Should return key when not found
	result := bundle.GetTranslation(LangEN, "unknown.key")
	assert.Equal(t, "unknown.key", result)
}

func TestBundle_HasTranslation(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")

	assert.True(t, bundle.HasTranslation(LangEN, "greeting"))
	assert.False(t, bundle.HasTranslation(LangEN, "unknown"))
	assert.False(t, bundle.HasTranslation(LangZH, "greeting"))
}

func TestBundle_Languages(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")
	bundle.AddTranslation(LangFR, "greeting", "Bonjour")

	langs := bundle.Languages()
	assert.Len(t, langs, 3)
	assert.Contains(t, langs, LangEN)
	assert.Contains(t, langs, LangZH)
	assert.Contains(t, langs, LangFR)
}

func TestBundle_Keys(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangEN, "farewell", "Goodbye")

	keys := bundle.Keys(LangEN)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "greeting")
	assert.Contains(t, keys, "farewell")

	// Empty for missing language
	keys = bundle.Keys(LangZH)
	assert.Nil(t, keys)
}

func TestBundle_Clear(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.Clear()

	assert.Empty(t, bundle.Languages())
}

func TestBundle_ClearLanguage(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")
	bundle.ClearLanguage(LangZH)

	assert.True(t, bundle.HasTranslation(LangEN, "greeting"))
	assert.False(t, bundle.HasTranslation(LangZH, "greeting"))
}

func TestBundle_LoadJSON(t *testing.T) {
	bundle := NewBundle(LangEN)
	jsonData := []byte(`{"greeting": "Hello", "farewell": "Goodbye"}`)

	err := bundle.LoadJSON(LangEN, jsonData)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "Goodbye", bundle.GetTranslation(LangEN, "farewell"))
}

func TestBundle_LoadJSON_Invalid(t *testing.T) {
	bundle := NewBundle(LangEN)
	jsonData := []byte(`invalid json`)

	err := bundle.LoadJSON(LangEN, jsonData)
	assert.Error(t, err)
}

func TestBundle_LoadYAML(t *testing.T) {
	bundle := NewBundle(LangEN)
	yamlData := []byte(`
greeting: Hello
farewell: Goodbye
`)

	err := bundle.LoadYAML(LangEN, yamlData)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "Goodbye", bundle.GetTranslation(LangEN, "farewell"))
}

func TestBundle_LoadYAML_Invalid(t *testing.T) {
	bundle := NewBundle(LangEN)
	yamlData := []byte(`invalid: [yaml`)

	err := bundle.LoadYAML(LangEN, yamlData)
	assert.Error(t, err)
}

func TestBundle_LoadJSONFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "en.json")
	err := os.WriteFile(filePath, []byte(`{"greeting": "Hello"}`), 0600)
	require.NoError(t, err)

	bundle := NewBundle(LangEN)
	err = bundle.LoadJSONFile(LangEN, filePath)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
}

func TestBundle_LoadJSONFile_NotFound(t *testing.T) {
	bundle := NewBundle(LangEN)
	err := bundle.LoadJSONFile(LangEN, "/nonexistent/file.json")
	assert.Error(t, err)
}

func TestBundle_LoadYAMLFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "en.yaml")
	err := os.WriteFile(filePath, []byte("greeting: Hello"), 0600)
	require.NoError(t, err)

	bundle := NewBundle(LangEN)
	err = bundle.LoadYAMLFile(LangEN, filePath)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
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

	bundle := NewBundle(LangEN)
	err = bundle.LoadDirectory(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "Hello", bundle.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "你好", bundle.GetTranslation(LangZH, "greeting"))
}

func TestBundle_LoadDirectory_NotFound(t *testing.T) {
	bundle := NewBundle(LangEN)
	err := bundle.LoadDirectory("/nonexistent/directory")
	assert.Error(t, err)
}

func TestBundle_Merge(t *testing.T) {
	bundle1 := NewBundle(LangEN)
	bundle1.AddTranslation(LangEN, "greeting", "Hello")
	bundle1.AddTranslation(LangEN, "farewell", "Goodbye")

	bundle2 := NewBundle(LangEN)
	bundle2.AddTranslation(LangEN, "greeting", "Hi") // Override
	bundle2.AddTranslation(LangZH, "greeting", "你好")

	bundle1.Merge(bundle2)

	assert.Equal(t, "Hi", bundle1.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "Goodbye", bundle1.GetTranslation(LangEN, "farewell"))
	assert.Equal(t, "你好", bundle1.GetTranslation(LangZH, "greeting"))
}

func TestBundle_Clone(t *testing.T) {
	original := NewBundle(LangEN)
	original.AddTranslation(LangEN, "greeting", "Hello")

	clone := original.Clone()

	// Clone should have same data
	assert.Equal(t, "Hello", clone.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, original.GetFallback(), clone.GetFallback())

	// Modifying clone should not affect original
	clone.AddTranslation(LangEN, "greeting", "Hi")
	assert.Equal(t, "Hello", original.GetTranslation(LangEN, "greeting"))
	assert.Equal(t, "Hi", clone.GetTranslation(LangEN, "greeting"))
}

func TestDefaultBundle(t *testing.T) {
	assert.NotNil(t, DefaultBundle)
	assert.Equal(t, LangEN, DefaultBundle.GetFallback())
}
