package i18n

import (
	"os"
	"path/filepath"
	"strings"
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

// TestLoadDirectory_JSONOnly is the root package's half of the v4 split: it
// loads .json and skips what it does not recognise, exactly as before.
func TestLoadDirectory_JSONOnly(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("en.json", `{"greeting": "Hello"}`)
	write("zh.json", `{"greeting": "你好"}`)
	write("unknown.json", `{"greeting": "?"}`) // not a language, skipped
	write("readme.txt", "readme")              // not a translation file, skipped

	bundle := NewBundle(LangEN)
	if err := bundle.LoadDirectory(dir); err != nil {
		t.Fatalf("LoadDirectory: %v", err)
	}

	if got := bundle.GetTranslation(LangEN, "greeting"); got != "Hello" {
		t.Errorf("en greeting = %q, want Hello", got)
	}
	if got := bundle.GetTranslation(LangZH, "greeting"); got != "你好" {
		t.Errorf("zh greeting = %q, want 你好", got)
	}
}

// TestLoadDirectory_RejectsYAMLLoudly pins the one behaviour change in v4.
//
// Skipping the YAML files would have been the quiet option and the wrong one:
// a directory of YAML translations would load as an empty bundle, every lookup
// would fall back to its key, and nobody would find out until a page of
// untranslated strings reached someone. The error names the fix.
func TestLoadDirectory_RejectsYAMLLoudly(t *testing.T) {
	for _, name := range []string{"zh.yaml", "zh.yml"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, name), []byte("greeting: 你好"), 0600); err != nil {
				t.Fatal(err)
			}

			err := NewBundle(LangEN).LoadDirectory(dir)
			if err == nil {
				t.Fatal("LoadDirectory returned nil for a directory of YAML, want an error")
			}
			if !strings.Contains(err.Error(), "yamlloader") {
				t.Errorf("error = %q, want it to name the yamlloader subpackage", err)
			}
		})
	}
}

// TestLoadDirectoryWith_CustomFormat is the extension point that lets
// yamlloader exist without this package importing a YAML library -- and lets
// anyone else add TOML or .properties on the same terms.
func TestLoadDirectoryWith_CustomFormat(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "en.txt"), []byte("greeting=Hello"), 0600); err != nil {
		t.Fatal(err)
	}

	lines := func(data []byte) (map[string]string, error) {
		k, v, _ := strings.Cut(string(data), "=")
		return map[string]string{k: v}, nil
	}

	bundle := NewBundle(LangEN)
	if err := bundle.LoadDirectoryWith(dir, map[string]Decoder{".txt": lines}); err != nil {
		t.Fatalf("LoadDirectoryWith: %v", err)
	}
	if got := bundle.GetTranslation(LangEN, "greeting"); got != "Hello" {
		t.Errorf("greeting = %q, want Hello", got)
	}
}

// TestLoadDirectoryWith_DecodeError reports which file failed, not just that
// something did.
func TestLoadDirectoryWith_DecodeError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "en.json"), []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}

	err := NewBundle(LangEN).LoadDirectory(dir)
	if err == nil {
		t.Fatal("LoadDirectory returned nil for malformed JSON, want an error")
	}
	if !strings.Contains(err.Error(), "en.json") {
		t.Errorf("error = %q, want it to name en.json", err)
	}
}
