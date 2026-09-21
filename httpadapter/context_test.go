package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	i18n "github.com/soulteary/i18n-kit/v4"
)

func TestSetLanguageInRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = SetLanguage(req, i18n.LangFR)

	lang := Language(req)
	assert.Equal(t, i18n.LangFR, lang)
}

func TestLanguageFromRequest_NoValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	lang := Language(req)
	assert.Equal(t, i18n.DefaultLanguage, lang)
}

func TestTFromRequest(t *testing.T) {
	defer i18n.DefaultBundle.Clear()

	i18n.DefaultBundle.AddTranslation(i18n.LangEN, "greeting", "Hello")
	i18n.DefaultBundle.AddTranslation(i18n.LangZH, "greeting", "你好")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = SetLanguage(req, i18n.LangZH)

	result := T(req, "greeting")
	assert.Equal(t, "你好", result)
}

func TestTfFromRequest(t *testing.T) {
	defer i18n.DefaultBundle.Clear()

	i18n.DefaultBundle.AddTranslation(i18n.LangZH, "greeting", "你好, %s!")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = SetLanguage(req, i18n.LangZH)

	result := Tf(req, "greeting", "世界")
	assert.Equal(t, "你好, 世界!", result)
}

func TestLanguage_Nil(t *testing.T) {
	lang := Language(nil)
	assert.Equal(t, i18n.DefaultLanguage, lang)
}

// TestNilRequestDoesNotPanic covers the three helpers that take an
// *http.Request and are reachable with a nil one -- Language already guards it,
// and T and Tf have to reach the same answer rather than dereferencing.
func TestNilRequestDoesNotPanic(t *testing.T) {
	defer i18n.DefaultBundle.Clear()
	i18n.DefaultBundle.AddTranslation(i18n.DefaultLanguage, "greeting", "Hello")

	assert.Equal(t, i18n.DefaultLanguage, Language(nil))
	assert.Equal(t, "Hello", T(nil, "greeting"))
	assert.Equal(t, "Hello", Tf(nil, "greeting"))
}
