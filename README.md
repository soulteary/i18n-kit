# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit/v2.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit/v2)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

[中文文档](README_CN.md)

A lightweight, flexible internationalization (i18n) library for Go applications. Supports language detection from HTTP requests, translation bundles, and middleware for both Fiber and net/http.

[中文文档](README_CN.md)

## Features

- **Multiple Language Support**: Built-in support for 10+ languages (EN, ZH, FR, DE, JA, KO, IT, ES, PT, RU)
- **Language Detection**: Automatic detection from query parameters, cookies, headers, and Accept-Language
- **Translation Bundles**: Thread-safe translation management with fallback support
- **Dual Framework Support**: Middleware for both Fiber and net/http
- **Context Integration**: Store and retrieve language from context
- **Named Parameters**: Support for `{name}` style parameter substitution
- **Pluralization**: Simple plural form handling
- **File Loading**: Load translations from JSON or YAML files
- **Zero Dependencies**: Only depends on Fiber for middleware (optional)

## Requirements

- **Go 1.27+** (`go.mod` declares `go 1.27.0`)
- `github.com/gofiber/fiber/v3` v3.4.0+ for the Fiber middleware

This v2 module line targets Fiber v3. Applications still on Fiber v2 should
remain on `github.com/soulteary/i18n-kit` v1.

## Installation

```bash
go get github.com/soulteary/i18n-kit/v2
```

Fiber integrations require Fiber v3.4.0 or later. Applications that still use Fiber v2 should remain on `github.com/soulteary/i18n-kit` v1.

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit/v2"
)

func main() {
    // Add translations
    i18n.AddTranslations(i18n.LangEN, map[string]string{
        "greeting": "Hello, World!",
        "farewell": "Goodbye!",
    })
    i18n.AddTranslations(i18n.LangZH, map[string]string{
        "greeting": "你好，世界！",
        "farewell": "再见！",
    })

    // Set global language
    i18n.SetGlobalLanguage(i18n.LangEN)
    fmt.Println(i18n.T("greeting")) // Output: Hello, World!

    i18n.SetGlobalLanguage(i18n.LangZH)
    fmt.Println(i18n.T("greeting")) // Output: 你好，世界！
}
```

### With net/http Middleware

```go
package main

import (
    "net/http"
    i18n "github.com/soulteary/i18n-kit/v2"
)

func main() {
    // Add translations
    i18n.AddTranslations(i18n.LangEN, map[string]string{
        "greeting": "Hello!",
    })
    i18n.AddTranslations(i18n.LangZH, map[string]string{
        "greeting": "你好！",
    })

    // Create handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get translation using language from request
        greeting := i18n.TFromRequest(r, "greeting")
        w.Write([]byte(greeting))
    })

    // Apply middleware
    http.Handle("/", i18n.StdMiddleware()(handler))
    http.ListenAndServe(":8080", nil)
}
```

Test with:
```bash
curl "http://localhost:8080/?lang=zh"  # Output: 你好！
curl "http://localhost:8080/?lang=en"  # Output: Hello!
```

### With Fiber Middleware

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    i18n "github.com/soulteary/i18n-kit/v2"
)

func main() {
    // Add translations
    i18n.AddTranslations(i18n.LangEN, map[string]string{
        "greeting": "Hello!",
    })
    i18n.AddTranslations(i18n.LangZH, map[string]string{
        "greeting": "你好！",
    })

    app := fiber.New()

    // Apply middleware
    app.Use(i18n.FiberMiddleware())

    app.Get("/", func(c fiber.Ctx) error {
        greeting := i18n.TFromFiber(c, "greeting")
        return c.SendString(greeting)
    })

    app.Listen(":8080")
}
```

## Language Detection

The library supports multiple detection methods with configurable priority:

```go
config := i18n.DetectorConfig{
    QueryParam:     "lang",           // Query parameter name
    CookieName:     "lang",           // Cookie name
    HeaderName:     "X-Language",     // Custom header name
    AcceptLanguage: true,             // Parse Accept-Language header
    Priority:       []string{"query", "cookie", "header", "accept"},
    Default:        i18n.LangEN,      // Fallback language
}

detector := i18n.NewDetector(config)
```

Detection priority (default order):
1. Query parameter (`?lang=zh`)
2. Cookie (`lang=zh`)
3. Custom header (`X-Language: zh`)
4. Accept-Language header (`Accept-Language: zh-CN,zh;q=0.9`)

## Translation Bundles

### Creating a Bundle

```go
bundle := i18n.NewBundle(i18n.LangEN) // English as fallback

// Add translations
bundle.AddTranslation(i18n.LangEN, "greeting", "Hello")
bundle.AddTranslation(i18n.LangZH, "greeting", "你好")

// Add multiple translations
bundle.AddTranslations(i18n.LangFR, map[string]string{
    "greeting": "Bonjour",
    "farewell": "Au revoir",
})
```

### Loading from Files

```go
// Load from JSON
bundle.LoadJSONFile(i18n.LangEN, "locales/en.json")

// Load from YAML
bundle.LoadYAMLFile(i18n.LangZH, "locales/zh.yaml")

// Load entire directory
// Files should be named: en.json, zh.yaml, fr.json, etc.
bundle.LoadDirectory("locales/")
```

### Managing a Bundle

```go
bundle.HasTranslation(i18n.LangEN, "welcome")
bundle.GetTranslation(i18n.LangEN, "welcome")      // returns the key when missing
bundle.LookupTranslation(i18n.LangEN, "welcome")   // (value, found)
bundle.Keys(i18n.LangEN)                           // every key for a language
bundle.Languages()                                 // every language with translations
bundle.GetFallback()
bundle.SetFallback(i18n.LangEN)

bundle.Merge(other)          // copy another bundle's translations in
clone := bundle.Clone()      // independent copy
bundle.ClearLanguage(i18n.LangFR)
bundle.Clear()
```

### Global Language

The package-level `T`/`Tf` read a process-wide language:

```go
i18n.SetGlobalLanguage(i18n.LangZH)
lang := i18n.GetGlobalLanguage()
```

Prefer the request- and context-scoped variants in a server; a global is a
single value shared by every concurrent request.

### Translation Fallback

When a translation is not found in the requested language, it falls back to the bundle's default language:

```go
bundle := i18n.NewBundle(i18n.LangEN)
bundle.AddTranslation(i18n.LangEN, "greeting", "Hello")

// No Chinese translation, falls back to English
result := bundle.GetTranslation(i18n.LangZH, "greeting")
// result == "Hello"

// No translation at all, returns the key
result = bundle.GetTranslation(i18n.LangEN, "unknown.key")
// result == "unknown.key"
```

## Named Parameters

Use `{name}` placeholders and pass a map:

```go
bundle.AddTranslation(i18n.LangEN, "welcome", "Welcome, {name}! You have {count} messages.")

formatter := i18n.NewFormatter(bundle)
result := formatter.Format(i18n.LangEN, "welcome", map[string]any{
    "name":  "Alice",
    "count": 5,
})
// "Welcome, Alice! You have 5 messages."

// Or package-level, against DefaultBundle
result = i18n.Format(i18n.LangEN, "welcome", map[string]any{"name": "Alice", "count": 5})
```

Substitution is a **single pass over the template**, which gives two guarantees
worth relying on:

- **A substituted value is never rescanned.** If `name` is `"{count}"`, it stays
  `"{count}"` in the output rather than being substituted again.
- **The result does not depend on map iteration order.** Go randomises that, so a
  value containing another placeholder used to produce different output on
  identical input between runs.

A placeholder with no matching parameter is **left in place**, not blanked — a
visible `{name}` in the output is a missing parameter, which is easier to notice
and to test for than an empty string. Literal braces around a placeholder are
preserved: `Hello {name}` inside `{"message":"Hello {name}"}` still substitutes
correctly.

Inspect a template before formatting:

```go
params := i18n.ExtractParams("Welcome, {name}!")         // ["name"]
has := i18n.HasParams("Welcome, {name}!")                // true
missing := i18n.ValidateParams(tmpl, providedParams)      // names with no value
```

`i18n.TemplatePattern` is the compiled placeholder regexp, exported for callers
that need to match it themselves.

## Printf-style Formatting

`Tf` and its variants treat the translation as a `fmt.Sprintf` format string:

```go
bundle.AddTranslation(i18n.LangEN, "greeting", "Hello, %s! You are %d.")

translator := i18n.NewTranslator(bundle)
translator.Tf("greeting", "Alice", 30)                  // "Hello, Alice! You are 30."
translator.TfWithLang(i18n.LangZH, "greeting", "Alice", 30)

// Package-level, against GlobalTranslator / DefaultBundle
i18n.Tf("greeting", "Alice", 30)
i18n.TfWithLang(i18n.LangZH, "greeting", "Alice", 30)

// Request- and context-scoped
i18n.TfFromRequest(r, "greeting", "Alice", 30)
i18n.TfFromContext(ctx, "greeting", "Alice", 30)
i18n.TfFromContextWithBundle(ctx, "greeting", "Alice", 30)
i18n.TfFromFiber(c, "greeting", "Alice", 30)
```

**When the translation is missing, the key is returned unformatted and the
arguments are discarded.** That matters because a missing key would otherwise be
used as the format string:

```go
// Key not present in any bundle:
i18n.Tf("error.account_locked", "user@example.com")
// returns "error.account_locked"
// NOT "error.account_locked%!(EXTRA string=user@example.com)"
```

Arguments to these calls are typically an email address, a phone number or a user
id, so that `%!(EXTRA …)` suffix put them inside the message shown to whoever
triggered the error.

A translation that **is** present is always passed through `fmt.Sprintf`, even
with no arguments, so its own printf escapes render: `"Save 10%%"` becomes
`"Save 10%"`.

Use `Bundle.LookupTranslation` when you need to know which case you are in:

```go
if value, ok := bundle.LookupTranslation(i18n.LangEN, key); ok {
    // a real translation
}
```

## Pluralization

Simple plural form support:

```go
bundle.AddTranslation(i18n.LangEN, "items.zero", "No items")
bundle.AddTranslation(i18n.LangEN, "items.one", "One item")
bundle.AddTranslation(i18n.LangEN, "items.other", "{count} items")

formatter := i18n.NewFormatter(bundle)

formatter.Pluralize(i18n.LangEN, "items", 0, nil)  // "No items"
formatter.Pluralize(i18n.LangEN, "items", 1, nil)  // "One item"
formatter.Pluralize(i18n.LangEN, "items", 5, nil)  // "5 items"

// Package-level, against DefaultBundle
i18n.Pluralize(i18n.LangEN, "items", 5, nil)

// No bundle needed for the two-form case
i18n.PluralizeSimple(5, "item", "items") // "items"
```

Keys are looked up as `<key>.zero`, `.one`, `.few`, `.many`, `.other`, matching
the fields of `i18n.PluralizationRule`. Extra named parameters are substituted as
usual, and `{count}` is always available.

## Context Integration

### With context.Context

```go
ctx := i18n.ContextWithLanguage(context.Background(), i18n.LangZH)

// Later in your code
lang := i18n.LanguageFromContext(ctx)
translation := i18n.TFromContext(ctx, "greeting")
```

### With http.Request

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Get language from request context
    lang := i18n.LanguageFromRequest(r)
    
    // Get translation
    greeting := i18n.TFromRequest(r, "greeting")
}
```

## Middleware Configuration

### Full Configuration

```go
config := i18n.MiddlewareConfig{
    Detector:       i18n.DefaultDetector,  // Language detector
    Bundle:         myBundle,              // Custom bundle (optional)
    SetCookie:      true,                  // Set language cookie
    CookieName:     "lang",
    CookieMaxAge:   86400 * 365,           // 1 year
    CookiePath:     "/",
    CookieSecure:   true,
    CookieHTTPOnly: true,
    CookieSameSite: "Lax",
    Next: func(c fiber.Ctx) bool {        // Skip middleware
        return c.Path() == "/health"
    },
}

app.Use(i18n.FiberMiddleware(config))
```

### Skip Middleware for Specific Paths

```go
// For Fiber
config := i18n.MiddlewareConfig{
    Next: func(c fiber.Ctx) bool {
        return c.Path() == "/api/internal"
    },
}

// For net/http
config := i18n.MiddlewareConfig{
    NextStd: func(r *http.Request) bool {
        return r.URL.Path == "/api/internal"
    },
}
```

## Supported Languages

Built-in language codes and their common variants:

| Language | Code | Variants |
|----------|------|----------|
| English | `en` | en-US, en-GB, en-AU |
| Chinese | `zh` | zh-CN, zh-TW, zh-Hans, zh-Hant |
| French | `fr` | fr-FR, fr-CA |
| German | `de` | de-DE, de-AT, de-CH |
| Japanese | `ja` | ja-JP |
| Korean | `ko` | ko-KR |
| Italian | `it` | it-IT |
| Spanish | `es` | es-ES, es-MX |
| Portuguese | `pt` | pt-PT, pt-BR |
| Russian | `ru` | ru-RU |

### Working with Language Codes

```go
lang := i18n.NormalizeLanguage("en-US")        // "en"
lang, ok := i18n.ParseLanguage("zh-Hans")      // ("zh", true)
i18n.LangEN.IsValid()                          // true
i18n.LangEN.String()                           // "en"
```

### Adding Custom Languages

```go
// Register a new language
i18n.RegisterLanguage(i18n.Language("ar"))

// Add aliases for the language
i18n.AddLanguageAlias("ar-SA", i18n.Language("ar"))
i18n.AddLanguageAlias("ar-EG", i18n.Language("ar"))
```

## Thread Safety

All components are thread-safe:
- `Bundle`: Safe for concurrent reads and writes
- `Translator`: Safe for concurrent use
- Global functions: Protected by mutex

## Upgrade Notes (v2.2.0)

One method was added; nothing was removed. Two behaviours change output.

- **A missing translation no longer leaks `Tf` arguments into the message.**
  `Tf` used the translation as a printf format string, and `GetTranslation`
  returns the *key* when no translation exists — so
  `Tf("error.account_locked", userEmail)` with that key missing produced
  `error.account_locked%!(EXTRA string=user@example.com)`, putting the email
  address in front of whoever triggered the error. Every `Tf` variant now returns
  the key unformatted when the translation was not found. **If you asserted on
  that `%!(EXTRA …)` output, those assertions change.**
- **Parameter substitution is a single pass.** It substituted one parameter at a
  time with `strings.ReplaceAll`, rescanning values it had already substituted —
  so a value containing another parameter's placeholder got substituted again,
  and because Go randomises map iteration order, *whether that happened varied
  between runs on identical input*. A substituted value is never rescanned now.
- **An unmatched placeholder is left in place** rather than silently blanked, so
  a missing parameter is visible in the output.
- **Literal braces no longer swallow a following placeholder.**
  `{"message":"Hello {name}"}` used to look up the nonexistent parameter
  `"message":"Hello {name` and leave the real `{name}` unsubstituted.
- **Package-level `Tf` follows `SetGlobalLanguage`.** It read
  `GlobalTranslator.GetLanguage()`, which `SetGlobalLanguage` never writes, while
  `T` read the global. After `SetGlobalLanguage(LangZH)`, `T` returned Chinese and
  `Tf` silently returned the English fallback.
- **A present translation is always formatted**, even with no arguments, so its
  own printf escapes render — `"Save 10%%"` is `"Save 10%"` again. Only a
  *missing* translation skips formatting.
- **`Bundle.LookupTranslation(lang, key) (string, bool)` is new**, for callers
  that need to tell a translation from a returned key.

## Best Practices

1. **Use a single Bundle per application**: Create one bundle and reuse it
2. **Load translations at startup**: Avoid loading during request handling
3. **Use context-based translations**: Prefer `TFromRequest`/`TFromContext` over global `T`
4. **Set fallback language**: Always configure a fallback for missing translations
5. **Use meaningful keys**: Use dot-notation like `error.not_found`, `page.home.title`
6. **Prefer named parameters over `Tf` for user-facing text**: `{name}` placeholders
   cannot consume arguments the way a printf format string can, and an unmatched
   one is visible rather than silent
7. **Check `LookupTranslation` in tests**: a key returned where a sentence was
   expected is the signal that a translation is missing

## Testing

```bash
go test ./...

# With coverage
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
