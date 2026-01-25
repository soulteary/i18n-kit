package i18n

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithLanguage(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithLanguage(ctx, LangZH)

	lang := LanguageFromContext(ctx)
	assert.Equal(t, LangZH, lang)
}

func TestLanguageFromContext_EmptyContext(t *testing.T) {
	//nolint:staticcheck // Testing nil context handling is intentional
	lang := LanguageFromContext(nil)
	assert.Equal(t, DefaultLanguage, lang)
}

func TestLanguageFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	lang := LanguageFromContext(ctx)
	assert.Equal(t, DefaultLanguage, lang)
}

func TestLanguageFromContextOK(t *testing.T) {
	// With language
	ctx := ContextWithLanguage(context.Background(), LangZH)
	lang, ok := LanguageFromContextOK(ctx)
	assert.True(t, ok)
	assert.Equal(t, LangZH, lang)

	// Without language
	ctx = context.Background()
	lang, ok = LanguageFromContextOK(ctx)
	assert.False(t, ok)
	assert.Empty(t, lang)

	// Nil context
	//nolint:staticcheck // Testing nil context handling is intentional
	lang, ok = LanguageFromContextOK(nil)
	assert.False(t, ok)
	assert.Empty(t, lang)
}

func TestSetLanguageInRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = SetLanguageInRequest(req, LangFR)

	lang := LanguageFromRequest(req)
	assert.Equal(t, LangFR, lang)
}

func TestLanguageFromRequest_Nil(t *testing.T) {
	lang := LanguageFromRequest(nil)
	assert.Equal(t, DefaultLanguage, lang)
}

func TestLanguageFromRequest_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	lang := LanguageFromRequest(req)
	assert.Equal(t, DefaultLanguage, lang)
}

func TestTFromContext(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello")
	DefaultBundle.AddTranslation(LangZH, "greeting", "你好")

	ctx := ContextWithLanguage(context.Background(), LangZH)
	result := TFromContext(ctx, "greeting")
	assert.Equal(t, "你好", result)
}

func TestTfFromContext(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangZH, "greeting", "你好, %s!")

	ctx := ContextWithLanguage(context.Background(), LangZH)
	result := TfFromContext(ctx, "greeting", "世界")
	assert.Equal(t, "你好, 世界!", result)
}

func TestTFromRequest(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello")
	DefaultBundle.AddTranslation(LangZH, "greeting", "你好")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = SetLanguageInRequest(req, LangZH)

	result := TFromRequest(req, "greeting")
	assert.Equal(t, "你好", result)
}

func TestTfFromRequest(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangZH, "greeting", "你好, %s!")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = SetLanguageInRequest(req, LangZH)

	result := TfFromRequest(req, "greeting", "世界")
	assert.Equal(t, "你好, 世界!", result)
}

func TestContextWithBundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Custom Hello")

	ctx := context.Background()
	ctx = ContextWithBundle(ctx, bundle)

	result := BundleFromContext(ctx)
	assert.Equal(t, bundle, result)
}

func TestBundleFromContext_EmptyContext(t *testing.T) {
	//nolint:staticcheck // Testing nil context handling is intentional
	bundle := BundleFromContext(nil)
	assert.Equal(t, DefaultBundle, bundle)
}

func TestBundleFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	bundle := BundleFromContext(ctx)
	assert.Equal(t, DefaultBundle, bundle)
}

func TestTFromContextWithBundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Custom Hello")
	bundle.AddTranslation(LangZH, "greeting", "自定义你好")

	ctx := context.Background()
	ctx = ContextWithBundle(ctx, bundle)
	ctx = ContextWithLanguage(ctx, LangZH)

	result := TFromContextWithBundle(ctx, "greeting")
	assert.Equal(t, "自定义你好", result)
}

func TestTfFromContextWithBundle(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangZH, "greeting", "自定义你好, %s!")

	ctx := context.Background()
	ctx = ContextWithBundle(ctx, bundle)
	ctx = ContextWithLanguage(ctx, LangZH)

	result := TfFromContextWithBundle(ctx, "greeting", "世界")
	assert.Equal(t, "自定义你好, 世界!", result)
}
