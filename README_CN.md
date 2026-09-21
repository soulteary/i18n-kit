# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit/v3.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit/v3)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

一个轻量级、灵活的 Go 国际化 (i18n) 库。支持从 HTTP 请求自动检测语言、翻译包管理，以及 Fiber 和 net/http 双框架中间件。

[English Documentation](README.md)


> **v3.0.0 —— Fiber 支持移入子包，模块路径升为 `/v3`。**
> Fiber 入口现位于 `github.com/soulteary/i18n-kit/v3/fiberadapter`，
> 于是导入根包不再把 Fiber（以及 fasthttp）链接进用不到它的二进制。
> 对一个 net/http 服务来说，这意味着**少链接 25 个包、go.sum 从 48 行降到 8 行、
> 二进制小 25%**（8080 KB → 6100 KB，实测程序只调用 `StdMiddleware` 和 `T()`）。
>
> 本次从根包删除了导出 API，因此作为新主版本发布，而不是 v2 的 minor：
> **v2.2.0 原样继续可用**，升级是你主动改 import path 的行为，
> 绝不会被 `go get -u` 悄悄换掉。
>
> | 原来 | 现在 |
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
> `MiddlewareConfig.Next` 也跟着搬了：一个 `func(fiber.Ctx) bool` 字段正是把
> Fiber 拖进根包的原因，现在它在 `fiberadapter.Config` 上，该结构体内嵌
> `i18n.MiddlewareConfig`。`NextStd` 与 net/http 一侧没有任何变化。
>
> 顺带修掉一个行为缺陷：`TfFromFiber` 把参数整个丢掉，于是 `"%s"` 在 Fiber 侧
> 原样输出，而 `TfFromContext` 是正常格式化的。`fiberadapter.Tf` 会格式化。
>
> **两个配置布尔字段被改名，使其零值就是文档中写的默认值。**
> 它们原本都是正向 bool，无法与「没设置」区分，于是各自打了个会误伤的补丁：
>
> | 原来 | 现在 | 原因 |
> |---|---|---|
> | `DetectorConfig.AcceptLanguage bool`（默认 `true`） | `DisableAcceptLanguage bool` | `NewDetector(DetectorConfig{})` 会保留默认 `Priority` 里的 `"accept"`，却把这一步关着，于是 Accept-Language 被静默忽略 |
> | `MiddlewareConfig.CookieHTTPOnly bool`（默认 `true`） | `DisableCookieHTTPOnly bool` | 只有在同时设了 `CookieName` 或 `CookieSameSite` 时才会采用你的值，于是「只改个 cookie 名字」会静默清掉 `HttpOnly` |
>
> 无论你原先设的是 `true` 还是 `false`，改名都会产生编译错误，而不是静默改变行为。
>
> **`CookieSameSite` 现在只在一处解析** —— `ResolveCookieSameSite` —— 不再每个框架各解析一遍。
> 大小写不敏感，`"disabled"` 表示不输出该属性，无法识别的值一律按 `"Lax"` 处理，
> `"None"` 会强制打开 `Secure`。此前 `"strict"` 在 Fiber 侧是 `Strict`、在 net/http 侧却是 `Lax`；
> 且 net/http 侧的 `SameSite=None` **不带 `Secure`**，会被浏览器直接丢弃 —— 那个 cookie 从未被存下来过。

## 特性

- **多语言支持**：内置支持 10+ 种语言（EN, ZH, FR, DE, JA, KO, IT, ES, PT, RU）
- **语言检测**：自动从查询参数、Cookie、Header 和 Accept-Language 检测语言
- **翻译包**：线程安全的翻译管理，支持回退机制
- **不绑定框架**：内置 net/http 中间件，Fiber v3 在子包中，其他任何框架约 20 行即可接入 —— 检测面向的是一个三方法接口
- **用到才付出代价**：根包只有一个非 stdlib 依赖（`gopkg.in/yaml.v3`）；Fiber 和 fasthttp 仅在你导入 `fiberadapter` 时才会被链接
- **上下文集成**：从 context 存取语言信息
- **命名参数**：支持 `{name}` 风格的参数替换
- **复数形式**：简单的复数形式处理
- **文件加载**：从 JSON 或 YAML 文件加载翻译

## 环境要求

- **Go 1.27+**（`go.mod` 声明 `go 1.27.0`）
- Fiber 中间件需要 `github.com/gofiber/fiber/v3` v3.4.0+

v3 模块线面向 Fiber v3，且只有 `fiberadapter` 子包会链接它。
仍在 Fiber v2 上的应用请继续使用 `github.com/soulteary/i18n-kit` v1。

## 安装

```bash
go get github.com/soulteary/i18n-kit/v3
```

Fiber 集成要求 Fiber v3.4.0 或更高版本。仍使用 Fiber v2 的应用应继续使用 `github.com/soulteary/i18n-kit` v1。

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit/v3"
)

func main() {
    // 添加翻译
    i18n.AddTranslations(i18n.LangEN, map[string]string{
        "greeting": "Hello, World!",
        "farewell": "Goodbye!",
    })
    i18n.AddTranslations(i18n.LangZH, map[string]string{
        "greeting": "你好，世界！",
        "farewell": "再见！",
    })

    // 设置全局语言
    i18n.SetGlobalLanguage(i18n.LangEN)
    fmt.Println(i18n.T("greeting")) // 输出: Hello, World!

    i18n.SetGlobalLanguage(i18n.LangZH)
    fmt.Println(i18n.T("greeting")) // 输出: 你好，世界！
}
```

### 使用 net/http 中间件

```go
package main

import (
    "net/http"
    i18n "github.com/soulteary/i18n-kit/v3"
)

func main() {
    // 添加翻译
    i18n.AddTranslations(i18n.LangEN, map[string]string{
        "greeting": "Hello!",
    })
    i18n.AddTranslations(i18n.LangZH, map[string]string{
        "greeting": "你好！",
    })

    // 创建处理器
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 使用请求中的语言获取翻译
        greeting := i18n.TFromRequest(r, "greeting")
        w.Write([]byte(greeting))
    })

    // 应用中间件
    http.Handle("/", i18n.StdMiddleware()(handler))
    http.ListenAndServe(":8080", nil)
}
```

测试：
```bash
curl "http://localhost:8080/?lang=zh"  # 输出: 你好！
curl "http://localhost:8080/?lang=en"  # 输出: Hello!
```

### 使用 Fiber 中间件

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    i18n "github.com/soulteary/i18n-kit/v3"
    "github.com/soulteary/i18n-kit/v3/fiberadapter"
)

func main() {
    // 添加翻译
    i18n.AddTranslations(i18n.LangEN, map[string]string{
        "greeting": "Hello!",
    })
    i18n.AddTranslations(i18n.LangZH, map[string]string{
        "greeting": "你好！",
    })

    app := fiber.New()

    // 应用中间件
    app.Use(fiberadapter.Middleware())

    app.Get("/", func(c fiber.Ctx) error {
        greeting := fiberadapter.T(c, "greeting")
        return c.SendString(greeting)
    })

    app.Listen(":8080")
}
```

## 语言检测

支持多种检测方式，可配置优先级：

```go
config := i18n.DetectorConfig{
    QueryParam: "lang",           // 查询参数名
    CookieName: "lang",           // Cookie 名
    HeaderName: "X-Language",     // 自定义 Header 名
    Priority:   []string{"query", "cookie", "header", "accept"},
    Default:    i18n.LangEN,      // 回退语言

    // DisableAcceptLanguage: true,  // 关闭 Accept-Language 解析
}

detector := i18n.NewDetector(config)
```

上面每个字段的零值都是可用的，因此 `i18n.DetectorConfig{}` 的行为与
`i18n.DefaultDetectorConfig()` 完全一致 —— 只写你要改的那几项即可。

检测优先级（默认顺序）：
1. 查询参数 (`?lang=zh`)
2. Cookie (`lang=zh`)
3. 自定义 Header (`X-Language: zh`)
4. Accept-Language Header (`Accept-Language: zh-CN,zh;q=0.9`)

从 `Priority` 中去掉某个方式即完全跳过它 —— 那一项根本不会被读取。
`DisableAcceptLanguage` 是最后一步的独立开关，所以你可以保留列表里的
`"accept"`、同时把解析关掉。

检测器面向的是"能回答三个问题"的任何东西，不限于 `*http.Request`：

```go
lang := detector.DetectFromRequest(r)             // net/http
lang := detector.Detect(i18n.RequestSourceOf(r))  // 同一件事的显式写法
lang := detector.Detect(mySource)                 // 任意 i18n.RequestSource
```

参见[适配其他框架](#适配其他框架)。

## 翻译包

### 创建翻译包

```go
bundle := i18n.NewBundle(i18n.LangEN) // 英语作为回退语言

// 添加单个翻译
bundle.AddTranslation(i18n.LangEN, "greeting", "Hello")
bundle.AddTranslation(i18n.LangZH, "greeting", "你好")

// 批量添加翻译
bundle.AddTranslations(i18n.LangFR, map[string]string{
    "greeting": "Bonjour",
    "farewell": "Au revoir",
})
```

### 从文件加载

```go
// 从 JSON 加载
bundle.LoadJSONFile(i18n.LangEN, "locales/en.json")

// 从 YAML 加载
bundle.LoadYAMLFile(i18n.LangZH, "locales/zh.yaml")

// 加载整个目录
// 文件命名: en.json, zh.yaml, fr.json 等
bundle.LoadDirectory("locales/")
```

### 管理翻译包

```go
bundle.HasTranslation(i18n.LangEN, "welcome")
bundle.GetTranslation(i18n.LangEN, "welcome")      // 缺失时返回 key 本身
bundle.LookupTranslation(i18n.LangEN, "welcome")   // (value, found)
bundle.Keys(i18n.LangEN)                           // 某语言的所有 key
bundle.Languages()                                 // 所有已有译文的语言
bundle.GetFallback()
bundle.SetFallback(i18n.LangEN)

bundle.Merge(other)          // 把另一个 bundle 的译文并进来
clone := bundle.Clone()      // 独立副本
bundle.ClearLanguage(i18n.LangFR)
bundle.Clear()
```

### 全局语言

包级的 `T`/`Tf` 读取一个进程级的语言设置：

```go
i18n.SetGlobalLanguage(i18n.LangZH)
lang := i18n.GetGlobalLanguage()
```

在服务端请优先使用请求级和上下文级的变体；全局变量是被所有并发请求共享的单一值。

### 翻译回退

当请求的语言中找不到翻译时，会回退到默认语言：

```go
bundle := i18n.NewBundle(i18n.LangEN)
bundle.AddTranslation(i18n.LangEN, "greeting", "Hello")

// 没有中文翻译，回退到英语
result := bundle.GetTranslation(i18n.LangZH, "greeting")
// result == "Hello"

// 完全没有翻译，返回键名
result = bundle.GetTranslation(i18n.LangEN, "unknown.key")
// result == "unknown.key"
```

## 命名参数

使用 `{name}` 形式的占位符，并传入一个 map：

```go
bundle.AddTranslation(i18n.LangEN, "welcome", "Welcome, {name}! You have {count} messages.")

formatter := i18n.NewFormatter(bundle)
result := formatter.Format(i18n.LangEN, "welcome", map[string]any{
    "name":  "Alice",
    "count": 5,
})
// "Welcome, Alice! You have 5 messages."

// 也可以用包级函数，作用于 DefaultBundle
result = i18n.Format(i18n.LangEN, "welcome", map[string]any{"name": "Alice", "count": 5})
```

替换过程是**对模板的单次遍历**，由此得到两个可以依赖的保证：

- **已替换进去的值不会被再次扫描。** 如果 `name` 的值是 `"{count}"`，输出里它就保持
  `"{count}"`，不会被再替换一轮。
- **结果不依赖 map 的遍历顺序。** Go 会随机化遍历顺序，所以此前"值里含有另一个占位符"
  的情况，在完全相同的输入上每次运行可能产出不同结果。

没有对应参数的占位符会**原样保留**，不会被清空——输出里看得见的 `{name}` 说明缺了一个
参数，这比留下空字符串更容易被发现、也更容易写断言。占位符外层的字面花括号会被保留：
`{"message":"Hello {name}"}` 里的 `Hello {name}` 仍能正确替换。

格式化之前可以先检查模板：

```go
params := i18n.ExtractParams("Welcome, {name}!")         // ["name"]
has := i18n.HasParams("Welcome, {name}!")                // true
missing := i18n.ValidateParams(tmpl, providedParams)      // 没有取到值的参数名
```

`i18n.TemplatePattern` 是编译好的占位符正则，供需要自己匹配的调用方使用。

## printf 风格格式化

`Tf` 及其变体把译文当作 `fmt.Sprintf` 的格式串：

```go
bundle.AddTranslation(i18n.LangEN, "greeting", "Hello, %s! You are %d.")

translator := i18n.NewTranslator(bundle)
translator.Tf("greeting", "Alice", 30)                  // "Hello, Alice! You are 30."
translator.TfWithLang(i18n.LangZH, "greeting", "Alice", 30)

// 包级函数，作用于 GlobalTranslator / DefaultBundle
i18n.Tf("greeting", "Alice", 30)
i18n.TfWithLang(i18n.LangZH, "greeting", "Alice", 30)

// 请求级与上下文级
i18n.TfFromRequest(r, "greeting", "Alice", 30)
i18n.TfFromContext(ctx, "greeting", "Alice", 30)
i18n.TfFromContextWithBundle(ctx, "greeting", "Alice", 30)
fiberadapter.Tf(c, "greeting", "Alice", 30)
```

**译文缺失时返回未格式化的 key，并丢弃参数。** 这一点很重要，否则缺失的 key 会被当作
格式串使用：

```go
// key 在任何 bundle 里都不存在：
i18n.Tf("error.account_locked", "user@example.com")
// 返回 "error.account_locked"
// 而不是 "error.account_locked%!(EXTRA string=user@example.com)"
```

这类调用的参数通常是邮箱地址、手机号或用户 ID，而那个 `%!(EXTRA …)` 后缀会把它们塞进
展示给触发该错误的人看的消息里。

**存在**的译文始终会经过 `fmt.Sprintf`，即便没有参数，这样它自带的 printf 转义才能
正常渲染：`"Save 10%%"` 会变成 `"Save 10%"`。

需要区分这两种情况时，请用 `Bundle.LookupTranslation`：

```go
if value, ok := bundle.LookupTranslation(i18n.LangEN, key); ok {
    // 这是一条真正的译文
}
```

## 复数形式

简单的复数形式支持：

```go
bundle.AddTranslation(i18n.LangEN, "items.zero", "No items")
bundle.AddTranslation(i18n.LangEN, "items.one", "One item")
bundle.AddTranslation(i18n.LangEN, "items.other", "{count} items")

formatter := i18n.NewFormatter(bundle)

formatter.Pluralize(i18n.LangEN, "items", 0, nil)  // "No items"
formatter.Pluralize(i18n.LangEN, "items", 1, nil)  // "One item"
formatter.Pluralize(i18n.LangEN, "items", 5, nil)  // "5 items"

// 包级函数，作用于 DefaultBundle
i18n.Pluralize(i18n.LangEN, "items", 5, nil)

// 只有单复数两种形态时不需要 bundle
i18n.PluralizeSimple(5, "item", "items") // "items"
```

key 按 `<key>.zero`、`.one`、`.few`、`.many`、`.other` 查找，与
`i18n.PluralizationRule` 的字段一一对应。额外的命名参数会照常替换，`{count}` 始终可用。

## 上下文集成

### 使用 context.Context

```go
ctx := i18n.ContextWithLanguage(context.Background(), i18n.LangZH)

// 后续代码中
lang := i18n.LanguageFromContext(ctx)
translation := i18n.TFromContext(ctx, "greeting")
```

### 使用 http.Request

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // 从请求上下文获取语言
    lang := i18n.LanguageFromRequest(r)
    
    // 获取翻译
    greeting := i18n.TFromRequest(r, "greeting")
}
```

## 中间件配置

### 完整配置

```go
config := fiberadapter.Config{
    MiddlewareConfig: i18n.MiddlewareConfig{
        Detector:       i18n.DefaultDetector,  // 语言检测器
        Bundle:         myBundle,              // 自定义翻译包（可选）
        SetCookie:      true,                  // 设置语言 Cookie
        CookieName:     "lang",
        CookieMaxAge:   86400 * 365,           // 1 年
        CookiePath:     "/",
        CookieSecure:   true,
        CookieSameSite: "Lax",           // Lax | Strict | None | disabled
    },
    Next: func(c fiber.Ctx) bool {             // 跳过中间件
        return c.Path() == "/health"
    },
}

app.Use(fiberadapter.Middleware(config))
```

与 `DetectorConfig` 一样，每个字段的零值都可用：`SetCookie` 默认关闭；
开启后 cookie 默认带 `HttpOnly`，除非你设置 `DisableCookieHTTPOnly`。

`CookieSameSite` 只在一处解析 —— `i18n.ResolveCookieSameSite` —— 因此所有框架读到的含义一致：

| 取值 | 结果 |
|---|---|
| `"Lax"`（默认） | `SameSite=Lax` |
| `"Strict"` | `SameSite=Strict` |
| `"None"` | `SameSite=None`，并强制打开 `Secure` |
| `"disabled"` | 完全不输出 `SameSite` 属性 |
| 其他任意值 | `Lax` |

匹配不区分大小写，首尾空白会被忽略。`"None"` 之所以强制 `Secure`，是因为浏览器
拒收不带 `Secure` 的该组合 —— 那个 cookie 根本不会被存下来。

### 跳过特定路径

```go
// Fiber
config := fiberadapter.Config{
    Next: func(c fiber.Ctx) bool {
        return c.Path() == "/api/internal"
    },
}

// net/http
config := i18n.MiddlewareConfig{
    NextStd: func(r *http.Request) bool {
        return r.URL.Path == "/api/internal"
    },
}
```

## 适配其他框架

检测是一条链，架在一个三方法接口之上。因此为 Echo、Gin、chi 或任何其他框架写适配器，
要做的只是说明查询参数、cookie 和 header 分别从哪里取：

```go
type RequestSource interface {
    Query(name string) string
    Cookie(name string) string
    Header(name string) string
}
```

其余一切 —— 优先级顺序、配置默认值与合并规则、`SameSite` 规则、`Tf` 的缺失 key 规则 ——
都从根包读取而不是重新实现一遍，这样新适配器不会与内置的那些产生分歧：

| 你需要 | 调用 |
|---|---|
| 跑检测链 | `detector.Detect(src)` |
| 套用配置默认值与合并规则 | `i18n.ResolveMiddlewareConfig(cfg...)` |
| 解析 `CookieSameSite` | `i18n.ResolveCookieSameSite(value)` |
| 正确实现 `Tf` | `bundle.LookupTranslation(...)` + `i18n.FormatTranslation(...)` |
| 适配 `*http.Request` | `i18n.RequestSourceOf(r)` |
| 就"结果存哪里"达成一致 | `i18n.LocalsLanguageKey`、`i18n.LocalsBundleKey` |

一个完整的适配器，以 Echo 为例：

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

`fiberadapter` 是同样的结构，可以作为参考对照阅读。

> **框架的「每请求存储」不是 `context.Context`。**
> `LocalsLanguageKey` 和 `LocalsBundleKey` 是普通字符串，而本包的 context key
> 是未导出类型 —— 所以存在 `LocalsLanguageKey` 下的语言**不会**通过
> `i18n.LanguageFromContext` 读回来，`i18n.TFromContext` 会返回默认语言且不作任何提示。
> 请用你自己适配器的读取函数；若想让 context 系列函数也能看到，请自行调用
> `i18n.ContextWithLanguage`。

## 支持的语言

内置语言代码及其常见变体：

| 语言 | 代码 | 变体 |
|------|------|------|
| 英语 | `en` | en-US, en-GB, en-AU |
| 中文 | `zh` | zh-CN, zh-TW, zh-Hans, zh-Hant |
| 法语 | `fr` | fr-FR, fr-CA |
| 德语 | `de` | de-DE, de-AT, de-CH |
| 日语 | `ja` | ja-JP |
| 韩语 | `ko` | ko-KR |
| 意大利语 | `it` | it-IT |
| 西班牙语 | `es` | es-ES, es-MX |
| 葡萄牙语 | `pt` | pt-PT, pt-BR |
| 俄语 | `ru` | ru-RU |

### 处理语言代码

```go
lang := i18n.NormalizeLanguage("en-US")        // "en"
lang, ok := i18n.ParseLanguage("zh-Hans")      // ("zh", true)
i18n.LangEN.IsValid()                          // true
i18n.LangEN.String()                           // "en"
```

### 添加自定义语言

```go
// 注册新语言
i18n.RegisterLanguage(i18n.Language("ar"))

// 添加语言别名
i18n.AddLanguageAlias("ar-SA", i18n.Language("ar"))
i18n.AddLanguageAlias("ar-EG", i18n.Language("ar"))
```

## 线程安全

所有组件都是线程安全的：
- `Bundle`：支持并发读写
- `Translator`：支持并发使用
- 全局函数：使用互斥锁保护

## 升级说明（v3.0.0）

这是一份机械式清单；每条背后的原因见本文件顶部的说明。

1. **改 import path** 为 `github.com/soulteary/i18n-kit/v3`。正因为路径变了，
   v2.2.0 原样继续可用 —— 不会有任何东西把你意外升上来。
2. **把 Fiber 入口改指向** `github.com/soulteary/i18n-kit/v3/fiberadapter`
   （对照表见顶部说明）。net/http 一侧没有任何迁移。
3. **把 `MiddlewareConfig.Next` 移到** `fiberadapter.Config` 上，该结构体内嵌
   `i18n.MiddlewareConfig`。`NextStd` 不变。
4. **两个布尔字段改名**，留空不写时默认值都是对的：

   | 原来 | 现在 |
   |---|---|
   | `DetectorConfig.AcceptLanguage: true` | *（删掉这行即可 —— 这就是默认值）* |
   | `DetectorConfig.AcceptLanguage: false` | `DisableAcceptLanguage: true` |
   | `MiddlewareConfig.CookieHTTPOnly: true` | *（删掉这行即可 —— 这就是默认值）* |
   | `MiddlewareConfig.CookieHTTPOnly: false` | `DisableCookieHTTPOnly: true` |

   旧字段的任何写法都会产生编译错误，因此不会有静默的行为变化。

有两处行为会改变输出：

- **Fiber 侧的 `Tf` 现在会应用参数了。** `TfFromFiber` 原先把参数整个丢掉，于是
  `"%s"` 原样输出，而 `TfFromContext` 是正常格式化的。`fiberadapter.Tf` 会格式化。
  **如果你此前通过自行预格式化来绕开这个问题，请移除那段绕行代码。**
- **`CookieSameSite` 在各处的解析结果完全一致了。** 此前每个中间件各自解析字符串，
  于是 `"strict"` 在 Fiber 侧是 `Strict`、在 net/http 侧却是 `Lax`；且 net/http 侧发出的
  `SameSite=None` **不带 `Secure`** —— 这个组合会被浏览器拒收，那个 cookie 从未被存下来过。
  现在匹配不区分大小写，`"disabled"` 不输出该属性，`"None"` 强制 `Secure`。
  **如果你依赖过「小写值悄悄等于 `Lax`」这一行为，请改成明确写出你想要的值。**

## 升级说明（v2.2.0）

新增一个方法，没有删除任何东西。有两处行为会改变输出。

- **译文缺失时，`Tf` 的参数不再泄露进消息。** `Tf` 此前把译文当作 printf 格式串，而
  `GetTranslation` 在没有译文时返回的是 *key*——于是当 key 缺失时，
  `Tf("error.account_locked", userEmail)` 会产出
  `error.account_locked%!(EXTRA string=user@example.com)`，把邮箱地址摆在触发该错误的
  人面前。现在所有 `Tf` 变体在找不到译文时都返回未格式化的 key。**如果你对那段
  `%!(EXTRA …)` 输出写过断言，这些断言需要改。**
- **参数替换改为单次遍历。** 此前是用 `strings.ReplaceAll` 一个参数一个参数地替换，
  会重新扫描已经替换进去的值——于是"值里含有另一个参数占位符"的情况会被再替换一次，
  而由于 Go 随机化 map 遍历顺序，*在完全相同的输入上，这件事会不会发生每次运行都可能
  不同*。现在已替换的值绝不会被再次扫描。
- **未匹配的占位符会原样保留**，而不是被默默清空，这样缺参数在输出里看得见。
- **字面花括号不再吞掉紧随其后的占位符。** `{"message":"Hello {name}"}` 此前会去查一个
  并不存在的参数 `"message":"Hello {name`，并让真正的 `{name}` 没被替换。
- **包级 `Tf` 会跟随 `SetGlobalLanguage`。** 它此前读的是
  `GlobalTranslator.GetLanguage()`，而 `SetGlobalLanguage` 从不写这个值，`T` 读的却是
  全局变量。于是 `SetGlobalLanguage(LangZH)` 之后，`T` 返回中文而 `Tf` 静默返回英文
  回退文本。
- **存在的译文始终会被格式化**，即便没有参数，这样它自带的 printf 转义才能渲染——
  `"Save 10%%"` 又是 `"Save 10%"` 了。只有*缺失*的译文会跳过格式化。
- **新增 `Bundle.LookupTranslation(lang, key) (string, bool)`**，供需要区分"真正的
  译文"和"返回的 key"的调用方使用。

## 最佳实践

1. **每个应用使用一个 Bundle**：创建一个翻译包并复用
2. **启动时加载翻译**：避免在请求处理时加载
3. **使用基于上下文的翻译**：优先使用 `TFromRequest`/`TFromContext`
4. **设置回退语言**：始终配置缺失翻译的回退
5. **使用有意义的键名**：使用点号分隔，如 `error.not_found`、`page.home.title`

## 测试

```bash
go test ./...

# 带覆盖率
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

## 许可证

Apache License 2.0 —— 详见 [LICENSE](LICENSE)。
