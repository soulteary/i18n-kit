# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit/v3.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit/v3)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

[中文文档](README_CN.md)

A lightweight, flexible internationalization (i18n) library for Go applications. Supports language detection from HTTP requests, translation bundles, and middleware for both Fiber and net/http.

[中文文档](README_CN.md)


> **v3.0.0 — Fiber support moved to a subpackage, and the module is now `/v3`.**
> The Fiber entry points are now `github.com/soulteary/i18n-kit/v3/fiberadapter`,
> so importing the root package no longer links Fiber (and fasthttp) into
> binaries that never use it. In a net/http service that means **25 fewer
> linked packages, a go.sum shrinking from 48 lines to 8, and a 25% smaller
> binary** (8080 KB → 6100 KB, measured on a program that only calls
> `StdMiddleware` and `T()`).
>
> This removes exported API from the root package, so it goes out as a new
> major version rather than a v2 minor: **v2.2.0 keeps working untouched**, and
> upgrading is a deliberate edit of your import path, never something
> `go get -u` does to you.
>
> | Before | After |
> |---|---|
> | `i18n.FiberMiddleware(...)` | `fiberadapter.Middleware(...)` |
> | `i18n.SimpleFiberMiddleware()` | `fiberadapter.SimpleMiddleware()` |
> | `i18n.DetectFromFiber(c)` | `fiberadapter.Detect(c)` |
> | `detector.DetectFromFiber(c)` | `fiberadapter.DetectWith(detector, c)` |
> | `i18n.LanguageFromFiberLocals(c)` | `fiberadapter.Language(c)` |
> | `i18n.BundleFromFiberLocals(c)` | `fiberadapter.Bundle(c)` |
> | `i18n.TFromFiber(c, key)` | `fiberadapter.T(c, key)` |
> | `i18n.TfFromFiber(c, key, args...)` | `fiberadapter.Tf(c, key, args...)` |
>
> `MiddlewareConfig.Next` moved too: a `func(fiber.Ctx) bool` field is exactly
> what pulled Fiber into the root package, so it now lives on
> `fiberadapter.Config`, which embeds `i18n.MiddlewareConfig`. `NextStd` and
> everything on the net/http side are unchanged.
>
> One behaviour fix rides along: `TfFromFiber` discarded its arguments
> entirely, so `"%s"` came out literal on Fiber while `TfFromContext`
> formatted it correctly. `fiberadapter.Tf` formats.
>
> **Two config booleans were renamed so their zero value is the documented
> default.** Both were plain positive bools that could not be told apart from
> "not set", and both had a workaround that misfired:
>
> | Before | After | Why |
> |---|---|---|
> | `DetectorConfig.AcceptLanguage bool` (default `true`) | `DisableAcceptLanguage bool` | `NewDetector(DetectorConfig{})` kept `"accept"` in the default `Priority` while leaving the step off, so Accept-Language was silently ignored |
> | `MiddlewareConfig.CookieHTTPOnly bool` (default `true`) | `DisableCookieHTTPOnly bool` | the merge took your value only once `CookieName` or `CookieSameSite` was set, so naming the cookie and nothing else silently cleared `HttpOnly` |
>
> Both renames are compile errors rather than silent behaviour changes, whichever
> value you were setting.
>
> **`CookieSameSite` is now interpreted in one place** — `ResolveCookieSameSite`
> — instead of once per framework. Matching is case-insensitive, `"disabled"`
> omits the attribute, anything unrecognised means `"Lax"`, and `"None"` forces
> `Secure` on. Previously `"strict"` meant `Strict` on Fiber and `Lax` on
> net/http, and net/http emitted `SameSite=None` *without* `Secure`, which
> browsers reject — so that cookie was never stored.

## Features

- **Multiple Language Support**: Built-in support for 10+ languages (EN, ZH, FR, DE, JA, KO, IT, ES, PT, RU)
- **Language Detection**: Automatic detection from query parameters, cookies, headers, and Accept-Language
- **Translation Bundles**: Thread-safe translation management with fallback support
- **Framework-Agnostic**: net/http middleware built in, Fiber v3 in a subpackage, and any other framework in ~20 lines — detection runs against a three-method interface
- **Pay Only For What You Import**: the root package pulls in one non-stdlib dependency (`gopkg.in/yaml.v3`); Fiber and fasthttp are linked only if you import `fiberadapter`
- **Context Integration**: Store and retrieve language from context
- **Named Parameters**: Support for `{name}` style parameter substitution
- **Pluralization**: Simple plural form handling
- **File Loading**: Load translations from JSON or YAML files

## Requirements

- **Go 1.27+** (`go.mod` declares `go 1.27.0`)
- `github.com/gofiber/fiber/v3` v3.4.0+ for the Fiber middleware

This v3 module line targets Fiber v3, and only the `fiberadapter` subpackage
links it. Applications still on Fiber v2 should remain on
`github.com/soulteary/i18n-kit` v1.

## Installation

```bash
go get github.com/soulteary/i18n-kit/v3
```

Fiber integrations require Fiber v3.4.0 or later. Applications that still use Fiber v2 should remain on `github.com/soulteary/i18n-kit` v1.

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit/v3"
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
    i18n "github.com/soulteary/i18n-kit/v3"
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
    i18n "github.com/soulteary/i18n-kit/v3"
    "github.com/soulteary/i18n-kit/v3/fiberadapter"
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
    app.Use(fiberadapter.Middleware())

    app.Get("/", func(c fiber.Ctx) error {
        greeting := fiberadapter.T(c, "greeting")
        return c.SendString(greeting)
    })

    app.Listen(":8080")
}
```

## Language Detection

The library supports multiple detection methods with configurable priority:

```go
config := i18n.DetectorConfig{
    QueryParam: "lang",           // Query parameter name
    CookieName: "lang",           // Cookie name
    HeaderName: "X-Language",     // Custom header name
    Priority:   []string{"query", "cookie", "header", "accept"},
    Default:    i18n.LangEN,      // Fallback language

    // DisableAcceptLanguage: true,  // opt out of Accept-Language parsing
}

detector := i18n.NewDetector(config)
```

Every field above has a usable zero value, so `i18n.DetectorConfig{}` behaves
exactly as `i18n.DefaultDetectorConfig()` does — set only what you want to
change.

Detection priority (default order):
1. Query parameter (`?lang=zh`)
2. Cookie (`lang=zh`)
3. Custom header (`X-Language: zh`)
4. Accept-Language header (`Accept-Language: zh-CN,zh;q=0.9`)

Removing a method from `Priority` skips it entirely — it is never even read.
`DisableAcceptLanguage` is the separate switch for the last step, so you can
leave `"accept"` in the list and still turn the parsing off.

A detector runs against anything that can answer three questions, not just an
`*http.Request`:

```go
lang := detector.DetectFromRequest(r)             // net/http
lang := detector.Detect(i18n.RequestSourceOf(r))  // the same thing, spelled out
lang := detector.Detect(mySource)                 // any i18n.RequestSource
```

See [Adapting Another Framework](#adapting-another-framework).

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
fiberadapter.Tf(c, "greeting", "Alice", 30)
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
config := fiberadapter.Config{
    MiddlewareConfig: i18n.MiddlewareConfig{
        Detector:       i18n.DefaultDetector,  // Language detector
        Bundle:         myBundle,              // Custom bundle (optional)
        SetCookie:      true,                  // Set language cookie
        CookieName:     "lang",
        CookieMaxAge:   86400 * 365,           // 1 year
        CookiePath:     "/",
        CookieSecure:   true,
        CookieSameSite: "Lax",           // Lax | Strict | None | disabled
    },
    Next: func(c fiber.Ctx) bool {             // Skip middleware
        return c.Path() == "/health"
    },
}

app.Use(fiberadapter.Middleware(config))
```

Like `DetectorConfig`, every field has a usable zero value: `SetCookie` is off,
and when it is on the cookie is `HttpOnly` unless you set
`DisableCookieHTTPOnly`.

`CookieSameSite` is interpreted in exactly one place — `i18n.ResolveCookieSameSite`
— so every framework reads it the same way:

| Value | Result |
|---|---|
| `"Lax"` (default) | `SameSite=Lax` |
| `"Strict"` | `SameSite=Strict` |
| `"None"` | `SameSite=None`, and `Secure` is forced on |
| `"disabled"` | no `SameSite` attribute at all |
| anything else | `Lax` |

Matching is case-insensitive and surrounding whitespace is ignored. `"None"`
forces `Secure` because browsers reject the combination without it — the cookie
would simply never be stored.

### Skip Middleware for Specific Paths

```go
// For Fiber
config := fiberadapter.Config{
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

## Adapting Another Framework

Detection is one chain over a three-method interface, so an adapter for Echo,
Gin, chi or anything else only has to say where a query parameter, a cookie and
a header come from:

```go
type RequestSource interface {
    Query(name string) string
    Cookie(name string) string
    Header(name string) string
}
```

Everything else — the priority order, the config defaults, the `SameSite`
rules, the missing-key rule for `Tf` — is read from the root package rather
than restated, so a new adapter cannot drift from the built-in ones:

| What you need | What to call |
|---|---|
| Run the detection chain | `detector.Detect(src)` |
| Apply the config defaults and merge rules | `i18n.ResolveMiddlewareConfig(cfg...)` |
| Interpret `CookieSameSite` | `i18n.ResolveCookieSameSite(value)` |
| Implement `Tf` correctly | `bundle.LookupTranslation(...)` + `i18n.FormatTranslation(...)` |
| Adapt an `*http.Request` | `i18n.RequestSourceOf(r)` |
| Agree on where to store the result | `i18n.LocalsLanguageKey`, `i18n.LocalsBundleKey` |

A complete adapter, for Echo:

```go
package echoadapter

import (
    "net/http"

    "github.com/labstack/echo/v4"
    i18n "github.com/soulteary/i18n-kit/v3"
)

type Source struct{ C echo.Context }

func (s Source) Query(name string) string { return s.C.QueryParam(name) }

func (s Source) Cookie(name string) string {
    cookie, err := s.C.Cookie(name)
    if err != nil {
        return ""
    }
    return cookie.Value
}

func (s Source) Header(name string) string { return s.C.Request().Header.Get(name) }

func Middleware(config ...i18n.MiddlewareConfig) echo.MiddlewareFunc {
    cfg := i18n.ResolveMiddlewareConfig(config...)

    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            lang := cfg.Detector.Detect(Source{C: c})
            c.Set(i18n.LocalsLanguageKey, lang)

            if cfg.Bundle != nil {
                c.Set(i18n.LocalsBundleKey, cfg.Bundle)
            }

            if cfg.SetCookie {
                mode := i18n.ResolveCookieSameSite(cfg.CookieSameSite)
                cookie := &http.Cookie{
                    Name:     cfg.CookieName,
                    Value:    string(lang),
                    MaxAge:   cfg.CookieMaxAge,
                    Path:     cfg.CookiePath,
                    Secure:   cfg.CookieSecure || mode.RequiresSecure(),
                    HttpOnly: !cfg.DisableCookieHTTPOnly,
                }
                if sameSite, ok := mode.HTTPSameSite(); ok {
                    cookie.SameSite = sameSite
                }
                c.SetCookie(cookie)
            }

            return next(c)
        }
    }
}

func Language(c echo.Context) i18n.Language {
    if lang, ok := c.Get(i18n.LocalsLanguageKey).(i18n.Language); ok {
        return lang
    }
    return i18n.DefaultLanguage
}

func Bundle(c echo.Context) *i18n.Bundle {
    if bundle, ok := c.Get(i18n.LocalsBundleKey).(*i18n.Bundle); ok {
        return bundle
    }
    return i18n.DefaultBundle
}

func T(c echo.Context, key string) string {
    return Bundle(c).GetTranslation(Language(c), key)
}

func Tf(c echo.Context, key string, args ...interface{}) string {
    text, found := Bundle(c).LookupTranslation(Language(c), key)
    return i18n.FormatTranslation(text, found, args...)
}
```

`fiberadapter` is the same shape and is worth reading as a reference.

> **A framework's per-request store is not a `context.Context`.**
> `LocalsLanguageKey` and `LocalsBundleKey` are plain strings, while this
> package's context keys have an unexported type — so a language stored under
> `LocalsLanguageKey` does **not** read back through `i18n.LanguageFromContext`,
> and `i18n.TFromContext` would answer in the default language without
> reporting anything. Read it back with your adapter's own accessor, or call
> `i18n.ContextWithLanguage` yourself if you want the context helpers to see it.

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

## Upgrade Notes (v3.0.0)

A mechanical checklist; the notice at the top of this file explains the
reasoning behind each one.

1. **Change the import path** to `github.com/soulteary/i18n-kit/v3`. Because
   the path changed, v2.2.0 keeps working untouched — nothing upgrades you by
   accident.
2. **Re-point the Fiber entry points** at
   `github.com/soulteary/i18n-kit/v3/fiberadapter` (see the table in the notice
   above). Nothing on the net/http side moved.
3. **Move `MiddlewareConfig.Next`** onto `fiberadapter.Config`, which embeds
   `i18n.MiddlewareConfig`. `NextStd` is unchanged.
4. **Rename two booleans**, both of which now default correctly when left alone:

   | Before | After |
   |---|---|
   | `DetectorConfig.AcceptLanguage: true` | *(delete the line — that is the default)* |
   | `DetectorConfig.AcceptLanguage: false` | `DisableAcceptLanguage: true` |
   | `MiddlewareConfig.CookieHTTPOnly: true` | *(delete the line — that is the default)* |
   | `MiddlewareConfig.CookieHTTPOnly: false` | `DisableCookieHTTPOnly: true` |

   Every spelling of the old fields is a compile error, so nothing changes
   behaviour silently.

Two behaviours change output:

- **`Tf` on Fiber now applies its arguments.** `TfFromFiber` discarded them
  entirely, so `"%s"` came out literal while `TfFromContext` formatted it
  correctly. `fiberadapter.Tf` formats. **If you worked around this by
  pre-formatting the string yourself, remove the workaround.**
- **`CookieSameSite` is interpreted identically everywhere.** Previously each
  middleware read the string itself, so `"strict"` meant `Strict` on Fiber and
  `Lax` on net/http, and net/http emitted `SameSite=None` *without* `Secure` —
  a combination browsers reject, so that cookie was never stored. Matching is
  now case-insensitive, `"disabled"` omits the attribute, and `"None"` forces
  `Secure`. **If you relied on a lowercase value quietly meaning `Lax`, spell
  out what you want.**

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
