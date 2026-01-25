package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFormatter(t *testing.T) {
	bundle := NewBundle(LangEN)
	formatter := NewFormatter(bundle)

	assert.NotNil(t, formatter)
}

func TestNewFormatter_NilBundle(t *testing.T) {
	formatter := NewFormatter(nil)
	assert.NotNil(t, formatter)
}

func TestFormatter_Format(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello, {name}!")
	bundle.AddTranslation(LangZH, "greeting", "你好, {name}!")

	formatter := NewFormatter(bundle)

	result := formatter.Format(LangEN, "greeting", map[string]interface{}{
		"name": "World",
	})
	assert.Equal(t, "Hello, World!", result)

	result = formatter.Format(LangZH, "greeting", map[string]interface{}{
		"name": "世界",
	})
	assert.Equal(t, "你好, 世界!", result)
}

func TestFormatter_Format_MultipleParams(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "message", "{user} sent {count} messages to {recipient}")

	formatter := NewFormatter(bundle)

	result := formatter.Format(LangEN, "message", map[string]interface{}{
		"user":      "Alice",
		"count":     5,
		"recipient": "Bob",
	})
	assert.Equal(t, "Alice sent 5 messages to Bob", result)
}

func TestFormatter_Format_NoParams(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "greeting", "Hello, {name}!")

	formatter := NewFormatter(bundle)

	result := formatter.Format(LangEN, "greeting", nil)
	assert.Equal(t, "Hello, {name}!", result)

	result = formatter.Format(LangEN, "greeting", map[string]interface{}{})
	assert.Equal(t, "Hello, {name}!", result)
}

func TestFormatter_Pluralize(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "items.zero", "No items")
	bundle.AddTranslation(LangEN, "items.one", "One item")
	bundle.AddTranslation(LangEN, "items.other", "{count} items")

	formatter := NewFormatter(bundle)

	result := formatter.Pluralize(LangEN, "items", 0, nil)
	assert.Equal(t, "No items", result)

	result = formatter.Pluralize(LangEN, "items", 1, nil)
	assert.Equal(t, "One item", result)

	result = formatter.Pluralize(LangEN, "items", 5, nil)
	assert.Equal(t, "5 items", result)
}

func TestFormatter_Pluralize_FallbackToBase(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "messages", "{count} messages")

	formatter := NewFormatter(bundle)

	// Should fallback to base key when .other doesn't exist
	result := formatter.Pluralize(LangEN, "messages", 5, nil)
	assert.Equal(t, "5 messages", result)
}

func TestFormatter_Pluralize_WithParams(t *testing.T) {
	bundle := NewBundle(LangEN)
	bundle.AddTranslation(LangEN, "user_items.other", "{user} has {count} items")

	formatter := NewFormatter(bundle)

	result := formatter.Pluralize(LangEN, "user_items", 3, map[string]interface{}{
		"user": "Alice",
	})
	assert.Equal(t, "Alice has 3 items", result)
}

func TestPluralizeSimple(t *testing.T) {
	assert.Equal(t, "item", PluralizeSimple(1, "item", "items"))
	assert.Equal(t, "items", PluralizeSimple(0, "item", "items"))
	assert.Equal(t, "items", PluralizeSimple(2, "item", "items"))
	assert.Equal(t, "items", PluralizeSimple(100, "item", "items"))
}

func TestFormat_Convenience(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangEN, "greeting", "Hello, {name}!")

	result := Format(LangEN, "greeting", map[string]interface{}{
		"name": "World",
	})
	assert.Equal(t, "Hello, World!", result)
}

func TestPluralize_Convenience(t *testing.T) {
	defer DefaultBundle.Clear()

	DefaultBundle.AddTranslation(LangEN, "items.one", "One item")
	DefaultBundle.AddTranslation(LangEN, "items.other", "{count} items")

	result := Pluralize(LangEN, "items", 1, nil)
	assert.Equal(t, "One item", result)

	result = Pluralize(LangEN, "items", 5, nil)
	assert.Equal(t, "5 items", result)
}

func TestExtractParams(t *testing.T) {
	tests := []struct {
		template string
		expected []string
	}{
		{"Hello, {name}!", []string{"name"}},
		{"{user} sent {count} to {recipient}", []string{"user", "count", "recipient"}},
		{"No params here", []string{}},
		{"{param} and {param} again", []string{"param"}}, // Deduped
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			result := ExtractParams(tt.template)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestHasParams(t *testing.T) {
	assert.True(t, HasParams("Hello, {name}!"))
	assert.True(t, HasParams("{a} and {b}"))
	assert.False(t, HasParams("No params"))
	assert.False(t, HasParams(""))
}

func TestValidateParams(t *testing.T) {
	template := "{user} sent {count} to {recipient}"

	// All params provided
	missing := ValidateParams(template, map[string]interface{}{
		"user":      "Alice",
		"count":     5,
		"recipient": "Bob",
	})
	assert.Empty(t, missing)

	// Missing some params
	missing = ValidateParams(template, map[string]interface{}{
		"user": "Alice",
	})
	assert.ElementsMatch(t, []string{"count", "recipient"}, missing)

	// No params provided
	missing = ValidateParams(template, nil)
	assert.Len(t, missing, 3)
}
