package i18n

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTranslator(t *testing.T) {
	bundle := NewBundle(LangEN)
	translator := NewTranslator(bundle)

	assert.NotNil(t, translator)
	assert.Equal(t, LangEN, translator.GetLanguage())
}

func TestNewTranslator_NilBundle(t *testing.T) {
	translator := NewTranslator(nil)
	assert.NotNil(t, translator)
	assert.Equal(t, DefaultBundle, translator.Bundle())
}

func TestNewTranslatorWithLanguage(t *testing.T) {
	bundle := NewBundle(LangEN)
	translator := NewTranslatorWithLanguage(bundle, LangZH)

	assert.Equal(t, LangZH, translator.GetLanguage())
}

func TestTranslator_SetLanguage(t *testing.T) {
	bundle := NewBundle(LangEN)
	translator := NewTranslator(bundle)

	translator.SetLanguage(LangZH)
	assert.Equal(t, LangZH, translator.GetLanguage())
}

func TestTranslator_T(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")

	translator := NewTranslator(bundle)

	translator.SetLanguage(LangEN)
	assert.Equal(t, "Hello", translator.T("greeting"))

	translator.SetLanguage(LangZH)
	assert.Equal(t, "你好", translator.T("greeting"))
}

func TestTranslator_Tf(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello, %s!")

	translator := NewTranslator(bundle)
	result := translator.Tf("greeting", "World")
	assert.Equal(t, "Hello, World!", result)
}

func TestTranslator_TWithLang(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")

	translator := NewTranslator(bundle)
	translator.SetLanguage(LangEN)

	// Should get Chinese even though current language is English
	assert.Equal(t, "你好", translator.TWithLang(LangZH, "greeting"))
}

func TestTranslator_TfWithLang(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangZH, "greeting", "你好, %s!")

	translator := NewTranslator(bundle)
	translator.SetLanguage(LangEN)

	result := translator.TfWithLang(LangZH, "greeting", "世界")
	assert.Equal(t, "你好, 世界!", result)
}

func TestTranslator_Bundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	translator := NewTranslator(bundle)

	assert.Equal(t, bundle, translator.Bundle())
}

func TestGlobalLanguage(t *testing.T) {
	// Save original state
	original := GetGlobalLanguage()
	defer SetGlobalLanguage(original)

	SetGlobalLanguage(LangZH)
	assert.Equal(t, LangZH, GetGlobalLanguage())

	SetGlobalLanguage(LangFR)
	assert.Equal(t, LangFR, GetGlobalLanguage())

	// Invalid language should not change
	SetGlobalLanguage(Language("invalid"))
	assert.Equal(t, LangFR, GetGlobalLanguage())
}

func TestGlobalT(t *testing.T) {
	// Save original state
	original := GetGlobalLanguage()
	defer func() {
		SetGlobalLanguage(original)
		DefaultBundle.Clear()
	}()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello")
	DefaultBundle.AddTranslation(LangZH, "greeting", "你好")

	SetGlobalLanguage(LangEN)
	assert.Equal(t, "Hello", T("greeting"))

	SetGlobalLanguage(LangZH)
	assert.Equal(t, "你好", T("greeting"))
}

func TestGlobalTf(t *testing.T) {
	// Save original state
	original := GetGlobalLanguage()
	defer func() {
		SetGlobalLanguage(original)
		DefaultBundle.Clear()
	}()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello, %s!")
	SetGlobalLanguage(LangEN)

	result := Tf("greeting", "World")
	assert.Equal(t, "Hello, World!", result)
}

func TestGlobalTWithLang(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello")
	DefaultBundle.AddTranslation(LangZH, "greeting", "你好")

	assert.Equal(t, "Hello", TWithLang(LangEN, "greeting"))
	assert.Equal(t, "你好", TWithLang(LangZH, "greeting"))
}

func TestGlobalTfWithLang(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangZH, "greeting", "你好, %s!")

	result := TfWithLang(LangZH, "greeting", "世界")
	assert.Equal(t, "你好, 世界!", result)
}

func TestAddTranslation(t *testing.T) {
	defer DefaultBundle.Clear()

	AddTranslation(LangEN, "test.key", "Test Value")
	assert.Equal(t, "Test Value", DefaultBundle.GetTranslation(LangEN, "test.key"))
}

func TestAddTranslations(t *testing.T) {
	defer DefaultBundle.Clear()

	AddTranslations(LangEN, map[string]string{
		"test.key1": "Value 1",
		"test.key2": "Value 2",
	})

	assert.Equal(t, "Value 1", DefaultBundle.GetTranslation(LangEN, "test.key1"))
	assert.Equal(t, "Value 2", DefaultBundle.GetTranslation(LangEN, "test.key2"))
}

func TestTranslator_ConcurrentAccess(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello")
	bundle.AddTranslation(LangZH, "greeting", "你好")

	translator := NewTranslator(bundle)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = translator.GetLanguage()
			_ = translator.T("greeting")
		}()
	}

	// Concurrent writes
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				translator.SetLanguage(LangEN)
			} else {
				translator.SetLanguage(LangZH)
			}
		}(i)
	}

	wg.Wait()

	// Should not panic and language should be valid
	lang := translator.GetLanguage()
	assert.True(t, lang == LangEN || lang == LangZH)
}

func TestGlobalTranslator_ConcurrentAccess(t *testing.T) {
	defer func() {
		DefaultBundle.Clear()
		SetGlobalLanguage(LangEN)
	}()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello")
	DefaultBundle.AddTranslation(LangZH, "greeting", "你好")

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent reads
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			_ = GetGlobalLanguage()
			_ = T("greeting")
		}()
	}

	// Concurrent writes
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetGlobalLanguage(LangEN)
			} else {
				SetGlobalLanguage(LangZH)
			}
		}(i)
	}

	wg.Wait()

	// Should not panic
	lang := GetGlobalLanguage()
	assert.True(t, lang == LangEN || lang == LangZH)
}
