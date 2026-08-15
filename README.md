# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)

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

## Installation

```bash
go get github.com/soulteary/i18n-kit
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit"
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
    i18n "github.com/soulteary/i18n-kit"
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
    "github.com/gofiber/fiber/v2"
    i18n "github.com/soulteary/i18n-kit"
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

    app.Get("/", func(c *fiber.Ctx) error {
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

Use `{name}` style placeholders for parameter substitution:

```go
bundle.AddTranslation(i18n.LangEN, "welcome", "Welcome, {name}! You have {count} messages.")

formatter := i18n.NewFormatter(bundle)
result := formatter.Format(i18n.LangEN, "welcome", map[string]interface{}{
    "name":  "Alice",
    "count": 5,
})
// result == "Welcome, Alice! You have 5 messages."
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
```

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
    Next: func(c *fiber.Ctx) bool {        // Skip middleware
        return c.Path() == "/health"
    },
}

app.Use(i18n.FiberMiddleware(config))
```

### Skip Middleware for Specific Paths

```go
// For Fiber
config := i18n.MiddlewareConfig{
    Next: func(c *fiber.Ctx) bool {
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

## Best Practices

1. **Use a single Bundle per application**: Create one bundle and reuse it
2. **Load translations at startup**: Avoid loading during request handling
3. **Use context-based translations**: Prefer `TFromRequest`/`TFromContext` over global `T`
4. **Set fallback language**: Always configure a fallback for missing translations
5. **Use meaningful keys**: Use dot-notation like `error.not_found`, `page.home.title`

## License

Apache License 2.0
