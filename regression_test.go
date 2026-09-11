package i18n

import (
	"context"
	"strings"
	"testing"
)

// TestMissingKeyDoesNotLeakArguments is the regression test for a missing
// translation turning into an information leak. GetTranslation returns the KEY
// when nothing is found, and Tf handed that key to fmt.Sprintf with the
// arguments, producing "error.locked%!(EXTRA string=user@example.com)" -- the
// argument, typically an email, phone number or user id, inside the message
// shown to whoever triggered it.
func TestMissingKeyDoesNotLeakArguments(t *testing.T) {
	b := NewBundle(LangEN)
	tr := NewTranslatorWithLanguage(b, LangEN)

	got := tr.Tf("error.account_locked", "user@example.com")

	if strings.Contains(got, "user@example.com") {
		t.Fatalf("Tf on a missing key leaked its argument: %q", got)
	}
	if strings.Contains(got, "EXTRA") {
		t.Errorf("Tf on a missing key produced fmt noise: %q", got)
	}
	if got != "error.account_locked" {
		t.Errorf("Tf on a missing key = %q, want the key itself", got)
	}

	// A present translation still formats.
	b.AddTranslation(LangEN, "greeting", "Hello %s")
	if got := tr.Tf("greeting", "world"); got != "Hello world" {
		t.Errorf("Tf on a present key = %q, want %q", got, "Hello world")
	}
}

func TestMissingKeyDoesNotLeakArgumentsAcrossEntryPoints(t *testing.T) {
	ctx := ContextWithLanguage(context.Background(), LangEN)

	for name, got := range map[string]string{
		"TfFromContext":           TfFromContext(ctx, "missing.key", "secret@example.com"),
		"TfFromContextWithBundle": TfFromContextWithBundle(ctx, "missing.key", "secret@example.com"),
		"TfWithLang":              NewTranslatorWithLanguage(NewBundle(LangEN), LangEN).TfWithLang(LangEN, "missing.key", "secret@example.com"),
	} {
		if strings.Contains(got, "secret@example.com") {
			t.Errorf("%s leaked its argument: %q", name, got)
		}
	}
}

// TestSubstituteParamsIsSinglePass: replacing one parameter at a time with
// strings.ReplaceAll rescanned substituted values, so a value containing
// another parameter's placeholder was substituted again -- and because Go
// randomises map iteration order, whether that happened varied run to run on
// identical input.
func TestSubstituteParamsIsSinglePass(t *testing.T) {
	template := "Hi {name}, your role is {role}"
	params := map[string]interface{}{
		"name": "{role}", // a value that looks like another placeholder
		"role": "admin",
	}

	want := "Hi {role}, your role is admin"

	// Run many times: the old implementation's outcome depended on map order.
	for i := 0; i < 200; i++ {
		if got := substituteParams(template, params); got != want {
			t.Fatalf("substituteParams = %q, want %q (iteration %d)", got, want, i)
		}
	}
}

// TestSubstituteParamsLeavesUnknownPlaceholders keeps a missing value visible
// instead of silently blank.
func TestSubstituteParamsLeavesUnknownPlaceholders(t *testing.T) {
	got := substituteParams("a {known} b {unknown} c", map[string]interface{}{"known": "X"})
	if got != "a X b {unknown} c" {
		t.Errorf("substituteParams = %q, want the unknown placeholder left in place", got)
	}
}

// TestLookupTranslationReportsMissing covers the flag the fix is built on.
func TestLookupTranslationReportsMissing(t *testing.T) {
	b := NewBundle(LangEN)
	b.AddTranslation(LangEN, "present", "Present")

	if text, found := b.LookupTranslation(LangEN, "present"); !found || text != "Present" {
		t.Errorf("LookupTranslation(present) = (%q, %v), want (\"Present\", true)", text, found)
	}
	if text, found := b.LookupTranslation(LangEN, "absent"); found || text != "absent" {
		t.Errorf("LookupTranslation(absent) = (%q, %v), want (\"absent\", false)", text, found)
	}
}

// --- Codex review follow-ups (PR #3) ---

// TestPackageTfFollowsSetGlobalLanguage is the regression test for package-level
// Tf reading GlobalTranslator's own language instead of globalLang. T reads
// globalLang and SetGlobalLanguage writes it, so after switching to Chinese, T
// returned Chinese while Tf silently fell back to English.
func TestPackageTfFollowsSetGlobalLanguage(t *testing.T) {
	const key = "codex.greeting"
	AddTranslation(LangEN, key, "Hello %s")
	AddTranslation(LangZH, key, "你好 %s")
	t.Cleanup(func() { SetGlobalLanguage(LangEN) })

	SetGlobalLanguage(LangZH)

	if got, want := T(key), "你好 %s"; got != want {
		t.Fatalf("T(%q) = %q, want %q", key, got, want)
	}
	if got, want := Tf(key, "世界"), "你好 世界"; got != want {
		t.Errorf("Tf(%q) = %q, want %q -- Tf must follow SetGlobalLanguage the way T does", key, got, want)
	}

	SetGlobalLanguage(LangEN)
	if got, want := Tf(key, "world"), "Hello world"; got != want {
		t.Errorf("Tf(%q) after switching back = %q, want %q", key, got, want)
	}
}

// TestPresentTranslationIsStillFormattedWithoutArgs: skipping fmt.Sprintf
// whenever the argument list was empty changed established output for
// translations carrying their own printf escapes. Only a MISSING translation
// may skip formatting -- that is what keeps arguments out of the message.
func TestPresentTranslationIsStillFormattedWithoutArgs(t *testing.T) {
	const key = "codex.discount"
	AddTranslation(LangEN, key, "Save 10%%")
	t.Cleanup(func() { SetGlobalLanguage(LangEN) })
	SetGlobalLanguage(LangEN)

	if got, want := Tf(key), "Save 10%"; got != want {
		t.Errorf("Tf(%q) = %q, want %q -- a present translation must still be formatted", key, got, want)
	}
	if got, want := TfWithLang(LangEN, key), "Save 10%"; got != want {
		t.Errorf("TfWithLang(%q) = %q, want %q", key, got, want)
	}
	if got, want := NewTranslatorWithLanguage(DefaultBundle, LangEN).Tf(key), "Save 10%"; got != want {
		t.Errorf("Translator.Tf(%q) = %q, want %q", key, got, want)
	}

	// The leak guard still holds: a missing key is returned bare.
	const missing = "codex.no.such.key"
	if got := Tf(missing, "user@example.com"); got != missing {
		t.Errorf("Tf(missing) = %q, want the bare key %q", got, missing)
	}
}
