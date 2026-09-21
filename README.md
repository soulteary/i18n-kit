# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit/v4.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit/v4)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A lightweight, flexible internationalization (i18n) library for Go applications. Supports language detection from HTTP requests, translation bundles, and middleware for both Fiber and net/http.

[中文文档](README_CN.md)


> **Breaking in v4.0.0 — new module path, and net/http and YAML each moved to a
> subpackage.**
>
> **Step 1 — everyone, including programs that serve no HTTP.** The module path
> is now `github.com/soulteary/i18n-kit/v4`:
>
> ```bash
> go get github.com/soulteary/i18n-kit/v4
> go mod edit -droprequire github.com/soulteary/i18n-kit/v3
> ```
>
> Then update the import path in your source. The major-version bump is required
> by Go's import compatibility rule, because v4 removes exported symbols;
> keeping them as shims was not an option, since a shim imports the very
> packages being moved.
>
> **Step 2 — net/http users.** The middleware, the `*http.Request` helpers and
> the request adapter moved to `github.com/soulteary/i18n-kit/v4/httpadapter`,
> so importing the root package no longer links a web server into a binary that
> never starts one. Translation is just as useful in a CLI printing localized
> help or error messages, and for a program importing only the root package this
> is **122 fewer linked packages and a 50% smaller binary** (203 → 81 packages,
> 3,940,615 → 1,974,432 bytes, `-trimpath -ldflags="-s -w"` on linux/amd64).
>
> | Before | After |
> |---|---|
> | `i18n.StdMiddleware(c)` | `httpadapter.Middleware(c)` |
> | `i18n.StdMiddlewareFunc(c)` | `httpadapter.MiddlewareFunc(c)` |
> | `i18n.SimpleMiddleware()` | `httpadapter.SimpleMiddleware()` |
> | `i18n.DetectFromRequest(r)` | `httpadapter.Detect(r)` |
> | `detector.DetectFromRequest(r)` | `httpadapter.DetectWith(detector, r)` |
> | `i18n.RequestSourceOf(r)` | `httpadapter.RequestSourceOf(r)` |
> | `i18n.SetLanguageInRequest(r, lang)` | `httpadapter.SetLanguage(r, lang)` |
> | `i18n.LanguageFromRequest(r)` | `httpadapter.Language(r)` |
> | `i18n.TFromRequest(r, key)` | `httpadapter.T(r, key)` |
> | `i18n.TfFromRequest(r, key, args...)` | `httpadapter.Tf(r, key, args...)` |
> | `mode.HTTPSameSite()` | `httpadapter.SameSite(mode)` |
> | `MiddlewareConfig.NextStd` | `httpadapter.Config.Next` |
>
> `MiddlewareConfig.NextStd` moved for the same reason `Next` moved to
> `fiberadapter.Config` in v3: a `func(*http.Request) bool` field is exactly what
> pulled net/http into the root package. `httpadapter.Config` embeds
> `i18n.MiddlewareConfig` and adds `Next`, mirroring `fiberadapter.Config`
> exactly.
>
> **Step 3 — YAML translation users.** `LoadYAML` and `LoadYAMLFile` moved to
> `github.com/soulteary/i18n-kit/v4/yamlloader`, because a YAML parser in the
> root package was paid for by every program, including the majority whose
> translation files are JSON — which the standard library already reads.
>
> | Before | After |
> |---|---|
> | `bundle.LoadYAML(lang, data)` | `yamlloader.Load(bundle, lang, data)` |
> | `bundle.LoadYAMLFile(lang, path)` | `yamlloader.LoadFile(bundle, lang, path)` |
> | `bundle.LoadDirectory(dir)` *(with any .yaml/.yml in it)* | `yamlloader.LoadDirectory(bundle, dir)` |
>
> **`Bundle.LoadDirectory` now reads only `.json`.** A directory containing
> `.yaml` or `.yml` returns an error naming `yamlloader.LoadDirectory` rather
> than silently loading half the translations — that is the one behaviour change
> in this release, and it is loud on purpose. An all-JSON directory is
> unaffected.
>
> **If you use neither, there is no step 2 or 3.** Bundles, the `Translator`,
> `Detector`, `RequestSource`, the context helpers, `MiddlewareConfig`,
> `ResolveMiddlewareConfig`, `ResolveCookieSameSite` and every formatting and
> pluralization function stayed in the root package with their v3 signatures.
> `fiberadapter` changes only its import path.
>
> **One root-package signature did change:** `Formatter.FormatWithContext` now
> takes a `context.Context` instead of an `interface{}`. It never worked through
> the old one -- the type switch it used matched nothing, so every call
> formatted in `DefaultLanguage` -- see step 5 of the upgrade notes.
>
> → **[Upgrade Notes (v4.0.0)](#upgrade-notes-v400)**

## Features

- **Multiple Language Support**: Built-in support for 10+ languages (EN, ZH, FR, DE, JA, KO, IT, ES, PT, RU)
- **Language Detection**: Automatic detection from query parameters, cookies, headers, and Accept-Language
- **Translation Bundles**: Thread-safe translation management with fallback support
- **Framework-Agnostic**: net/http in `httpadapter`, Fiber v3 in `fiberadapter`, and any other framework in ~20 lines — detection runs against a three-method interface
- **Pay Only For What You Import**: the root package depends on nothing outside the standard library and does not import `net/http` — the web server, Fiber, and the YAML parser each live behind their own subpackage
- **Context Integration**: Store and retrieve language from context
- **Named Parameters**: Support for `{name}` style parameter substitution
- **Pluralization**: Simple plural form handling
- **File Loading**: JSON out of the box; YAML via the `yamlloader` subpackage

## Requirements

- **Go 1.27+** (`go.mod` declares `go 1.27.0`)
- `github.com/gofiber/fiber/v3` v3.5.0+ — only for the `fiberadapter` subpackage
- `gopkg.in/yaml.v3` — only for the `yamlloader` subpackage
- the `httpadapter` subpackage and the root package need nothing but the standard library

This v4 module line targets Fiber v3, and only the `fiberadapter` subpackage
links it. Applications still on Fiber v2 should remain on
`github.com/soulteary/i18n-kit` v1.

## Installation

```bash
go get github.com/soulteary/i18n-kit/v4
```

The root package depends on nothing outside the standard library — not even
`net/http`. Everything that needs a dependency lives in its own subpackage, so a
binary links only what it actually uses:

```bash
# net/http middleware and *http.Request helpers — standard library only
go get github.com/soulteary/i18n-kit/v4/httpadapter

# YAML translation files — links gopkg.in/yaml.v3
go get github.com/soulteary/i18n-kit/v4/yamlloader

# Fiber v3 middleware — links Fiber, and with it fasthttp
go get github.com/soulteary/i18n-kit/v4/fiberadapter
```

Fiber integrations require Fiber v3.5.0 or later — the version `go.mod` requires. Applications that still use Fiber v2 should remain on `github.com/soulteary/i18n-kit` v1.

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit/v4"
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
    i18n "github.com/soulteary/i18n-kit/v4"
    "github.com/soulteary/i18n-kit/v4/httpadapter"
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
        greeting := httpadapter.T(r, "greeting")
        w.Write([]byte(greeting))
    })

    // Apply middleware
    http.Handle("/", httpadapter.Middleware()(handler))
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
    i18n "github.com/soulteary/i18n-kit/v4"
    "github.com/soulteary/i18n-kit/v4/fiberadapter"
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
lang := detector.Detect(httpadapter.RequestSourceOf(r))  // the same thing, spelled out
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
// JSON — standard library only, no extra dependency
bundle.LoadJSONFile(i18n.LangEN, "locales/en.json")
bundle.LoadJSON(i18n.LangEN, data)   // from bytes: an embed.FS, a response body

// YAML — links gopkg.in/yaml.v3, so it lives in the yamlloader subpackage
yamlloader.LoadFile(bundle, i18n.LangZH, "locales/zh.yaml")
yamlloader.Load(bundle, i18n.LangZH, data)

// A whole directory, its files named en.json, zh.yaml, fr.json, …
bundle.LoadDirectory("locales/")               // .json only
yamlloader.LoadDirectory(bundle, "locales/")   // .json, .yaml and .yml
```

`Bundle.LoadDirectory` reads **only `.json`**. A directory holding `.yaml` or
`.yml` returns an error naming `yamlloader.LoadDirectory` rather than silently
loading half the translations. `yamlloader.LoadDirectory` keeps `.json`
included, so renaming one file to `.yaml` does not stop the rest from loading.

#### Teaching the loader another format

A `Decoder` turns one file's bytes into a flat map, which is how the root
package loads formats it links no parser for:

```go
type Decoder func(data []byte) (map[string]string, error)
```

`LoadDirectoryWith` takes the extensions it should understand:

```go
bundle.LoadDirectoryWith("locales/", map[string]i18n.Decoder{
    ".json": i18n.DecodeJSON,   // what LoadDirectory uses
    ".yaml": yamlloader.Decode, // what yamlloader.LoadDirectory adds
    ".toml": myTOMLDecoder,     // anything else, same three-line shape
})
```

This is exactly how `yamlloader.LoadDirectory` supports YAML without the root
package importing a YAML parser — and the same seam is open to your own formats.

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

`i18n.TWithLang` / `i18n.TfWithLang` name the language per call and read no
global at all, which is the simplest way to avoid that sharing:

```go
i18n.TWithLang(i18n.LangZH, "greeting")              // "你好，世界！"
i18n.TfWithLang(i18n.LangZH, "greeting", "Alice")
```

### Scoped Translators

A `Translator` binds a bundle to a current language, so nothing is shared with
the rest of the process:

```go
translator := i18n.NewTranslator(bundle)                           // starts at DefaultLanguage
translator = i18n.NewTranslatorWithLanguage(bundle, i18n.LangZH)   // or start somewhere else

translator.SetLanguage(i18n.LangFR)
translator.GetLanguage()                        // "fr"
translator.Bundle()                             // the bundle it was built with

translator.T("greeting")                        // in the translator's language
translator.Tf("greeting", "Alice")
translator.TWithLang(i18n.LangZH, "greeting")   // overriding it for one call
translator.TfWithLang(i18n.LangZH, "greeting", "Alice")
```

`i18n.GlobalTranslator` is the package-level one backing `T`/`Tf`. It is a
shared value, so prefer your own translator, or the context helpers below, over
mutating it.

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

`i18n.Format` and `i18n.Pluralize` are thin wrappers over `i18n.DefaultFormatter`,
the `*Formatter` built on `DefaultBundle`. Construct your own with `NewFormatter`
whenever the translations are not the process-wide ones.

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
httpadapter.Tf(r, "greeting", "Alice", 30)
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
lang := i18n.LanguageFromContext(ctx)          // DefaultLanguage when absent
lang, ok := i18n.LanguageFromContextOK(ctx)    // ("", false) when absent
translation := i18n.TFromContext(ctx, "greeting")
formatted := i18n.TfFromContext(ctx, "greeting", "Alice")
```

Use `LanguageFromContextOK` when "no language in this context" has to be told
apart from "the default language" — `LanguageFromContext` collapses the two.

### Carrying a Bundle in the Context

The context can hold a bundle as well as a language, which is what a server
serving several tenants — each with its own translations — needs:

```go
ctx = i18n.ContextWithBundle(ctx, tenantBundle)

bundle := i18n.BundleFromContext(ctx)                  // DefaultBundle when absent
text := i18n.TFromContextWithBundle(ctx, "greeting")
text = i18n.TfFromContextWithBundle(ctx, "greeting", "Alice")
```

The `WithBundle` variants read the bundle from the context; the plain
`TFromContext` / `TfFromContext` always read `DefaultBundle`. `httpadapter`'s
middleware stores both values, so a handler's `r.Context()` carries whichever
bundle the config set.

A `Formatter` resolves the language from a context too:

```go
formatter.FormatWithContext(r.Context(), "welcome", map[string]any{"name": "Alice"})
```

It formats against the formatter's *own* bundle — `ContextWithBundle` does not
override that — and falls back to `DefaultLanguage` when the context carries no
language. Fiber keeps the language in `Locals` rather than in a context: read it
with `fiberadapter.Language(c)` and call `Format` directly.

### With http.Request

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Get language from request context
    lang := httpadapter.Language(r)
    
    // Get translation
    greeting := httpadapter.T(r, "greeting")
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

`ResolveCookieSameSite` returns an `i18n.CookieSameSiteMode` — one of
`SameSiteLax`, `SameSiteStrict`, `SameSiteNone` and `SameSiteDisabled` — and
`mode.RequiresSecure()` reports the `SameSiteNone` case an adapter has to honour.
`httpadapter.SameSite(mode)` converts it to the `http.SameSite` constant
net/http wants:

```go
mode := i18n.ResolveCookieSameSite(cfg.CookieSameSite)   // i18n.SameSiteLax
if mode.RequiresSecure() {
    cookie.Secure = true
}
sameSite, write := httpadapter.SameSite(mode)            // write is false for "disabled"
```

`i18n.DefaultMiddlewareConfig()` returns the defaults these fields fall back to,
and `i18n.ResolveMiddlewareConfig(cfg...)` is what every adapter runs a caller's
config through — use it when adapting a framework of your own, so the defaults
stay in one place.

### Skip Middleware for Specific Paths

```go
// For Fiber
config := fiberadapter.Config{
    Next: func(c fiber.Ctx) bool {
        return c.Path() == "/api/internal"
    },
}

// For net/http
config := httpadapter.Config{
    Next: func(r *http.Request) bool {
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
| Adapt an `*http.Request` | `httpadapter.RequestSourceOf(r)` |
| Agree on where to store the result | `i18n.LocalsLanguageKey`, `i18n.LocalsBundleKey` |

A complete adapter, for Echo:

```go
package echoadapter

import (
    "net/http"

    "github.com/labstack/echo/v4"
    i18n "github.com/soulteary/i18n-kit/v4"
    "github.com/soulteary/i18n-kit/v4/httpadapter"
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
                if sameSite, ok := httpadapter.SameSite(mode); ok {
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

i18n.SupportedLanguages                        // []Language of the built-ins above
```

`SupportedLanguages` is the slice `IsValid` checks against, and `RegisterLanguage`
appends to it — range over it to build a language switcher rather than hardcoding
the table above.

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

## Upgrade Notes (v4.0.0)

1. **Change the module path.** Every import, in every file:

   ```bash
   go get github.com/soulteary/i18n-kit/v4
   go mod edit -droprequire github.com/soulteary/i18n-kit/v3
   ```

   ```diff
   -i18n "github.com/soulteary/i18n-kit/v3"
   +i18n "github.com/soulteary/i18n-kit/v4"
   ```

   `go get -u` will not do this for you; v3 stays on `v3.0.0`.

2. **Re-point the net/http entry points** at
   `github.com/soulteary/i18n-kit/v4/httpadapter` — the table in the notice at
   the top of this file lists all eleven. Names lost the parts that the package
   name now carries: `StdMiddleware` is `httpadapter.Middleware`,
   `LanguageFromRequest` is `httpadapter.Language`, and so on, mirroring
   `fiberadapter`.

3. **Move `MiddlewareConfig.NextStd`** onto `httpadapter.Config`, which embeds
   `i18n.MiddlewareConfig`, exactly as `fiberadapter.Config` does for Fiber's
   `Next`:

   ```diff
   -config := i18n.MiddlewareConfig{
   -    NextStd: skip,
   -}
   -handler = i18n.StdMiddleware(config)(handler)
   +config := httpadapter.Config{
   +    MiddlewareConfig: i18n.MiddlewareConfig{ /* ... */ },
   +    Next:             skip,
   +}
   +handler = httpadapter.Middleware(config)(handler)
   ```

4. **Re-point YAML loading** at `github.com/soulteary/i18n-kit/v4/yamlloader`:

   ```diff
   -err := bundle.LoadYAMLFile(i18n.LangZH, "locales/zh.yaml")
   +err := yamlloader.LoadFile(bundle, i18n.LangZH, "locales/zh.yaml")
   ```

   **If you load a directory containing YAML, this one is not a compile error.**
   `Bundle.LoadDirectory` still exists and still compiles; it now reads only
   `.json` and returns an error when it meets a `.yaml` or `.yml` file, naming
   `yamlloader.LoadDirectory` as the fix. It was made loud rather than silent
   precisely because a half-loaded translation set is invisible until someone
   reports a page of untranslated keys. Directories of JSON are unaffected.

   `LoadDirectoryWith` is new and takes the extensions it understands as a
   `map[string]i18n.Decoder`, which is how `yamlloader.LoadDirectory` adds YAML
   without this package importing a YAML library. TOML or `.properties` is
   twenty lines and needs no change here.

5. **`Formatter.FormatWithContext` now takes a `context.Context`.** This one is
   a fix, and it applies whether or not you use net/http or YAML:

   ```diff
   -func (f *Formatter) FormatWithContext(ctx interface{}, key string, params map[string]interface{}) string
   +func (f *Formatter) FormatWithContext(ctx context.Context, key string, params map[string]interface{}) string
   ```

   It used to take an `interface{}` and type-switch on
   `interface{ Context() interface{} }` to recognise a Fiber context. **No type
   has ever satisfied that switch** — `fiber.Ctx` declares
   `Context() context.Context`, not `Context() interface{}` — so every call fell
   through to `DefaultLanguage`, including one handed a `context.Context` built
   by `ContextWithLanguage`. The method had no test of its own, which is how it
   stayed that way. It now resolves the language through `LanguageFromContext`:

   ```go
   // net/http — the middleware stores the language in the request context
   formatter.FormatWithContext(r.Context(), "welcome", params)

   // Fiber — the language lives in Locals, not in a context
   formatter.Format(fiberadapter.Language(c), "welcome", params)
   ```

   Code already passing a `context.Context` keeps compiling and starts getting
   the context's language instead of always the default. Passing anything else
   becomes a compile error — which is the point, since it silently returned the
   wrong answer before.

6. **If you use neither net/http nor YAML, steps 2 to 4 do not apply.** Bundles,
   `Translator`, `Detector`, `RequestSource`, the context helpers,
   `MiddlewareConfig`, `ResolveMiddlewareConfig`, `ResolveCookieSameSite`, and
   every formatting and pluralization function other than `FormatWithContext`
   kept their v3 signatures in the root package. `fiberadapter` users change
   only the import path.

Nothing else changed: no detection rule, no cookie attribute, no translation
output, and no formatting behaviour beyond the `FormatWithContext` fix above.

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
