# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit/v2.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit/v2)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)

一个轻量级、灵活的 Go 国际化 (i18n) 库。支持从 HTTP 请求自动检测语言、翻译包管理，以及 Fiber 和 net/http 双框架中间件。

[English Documentation](README.md)

## 特性

- **多语言支持**：内置支持 10+ 种语言（EN, ZH, FR, DE, JA, KO, IT, ES, PT, RU）
- **语言检测**：自动从查询参数、Cookie、Header 和 Accept-Language 检测语言
- **翻译包**：线程安全的翻译管理，支持回退机制
- **双框架支持**：同时支持 Fiber 和 net/http 中间件
- **上下文集成**：从 context 存取语言信息
- **命名参数**：支持 `{name}` 风格的参数替换
- **复数形式**：简单的复数形式处理
- **文件加载**：从 JSON 或 YAML 文件加载翻译
- **零依赖**：仅 Fiber 中间件可选依赖

## 安装

```bash
go get github.com/soulteary/i18n-kit/v2
```

Fiber 集成要求 Fiber v3.4.0 或更高版本。仍使用 Fiber v2 的应用应继续使用 `github.com/soulteary/i18n-kit` v1。

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit/v2"
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
    i18n "github.com/soulteary/i18n-kit/v2"
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
    i18n "github.com/soulteary/i18n-kit/v2"
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
    app.Use(i18n.FiberMiddleware())

    app.Get("/", func(c fiber.Ctx) error {
        greeting := i18n.TFromFiber(c, "greeting")
        return c.SendString(greeting)
    })

    app.Listen(":8080")
}
```

## 语言检测

支持多种检测方式，可配置优先级：

```go
config := i18n.DetectorConfig{
    QueryParam:     "lang",           // 查询参数名
    CookieName:     "lang",           // Cookie 名
    HeaderName:     "X-Language",     // 自定义 Header 名
    AcceptLanguage: true,             // 解析 Accept-Language
    Priority:       []string{"query", "cookie", "header", "accept"},
    Default:        i18n.LangEN,      // 回退语言
}

detector := i18n.NewDetector(config)
```

检测优先级（默认顺序）：
1. 查询参数 (`?lang=zh`)
2. Cookie (`lang=zh`)
3. 自定义 Header (`X-Language: zh`)
4. Accept-Language Header (`Accept-Language: zh-CN,zh;q=0.9`)

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

使用 `{name}` 风格的占位符进行参数替换：

```go
bundle.AddTranslation(i18n.LangEN, "welcome", "Welcome, {name}! You have {count} messages.")

formatter := i18n.NewFormatter(bundle)
result := formatter.Format(i18n.LangEN, "welcome", map[string]interface{}{
    "name":  "Alice",
    "count": 5,
})
// result == "Welcome, Alice! You have 5 messages."
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
```

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
config := i18n.MiddlewareConfig{
    Detector:       i18n.DefaultDetector,  // 语言检测器
    Bundle:         myBundle,              // 自定义翻译包（可选）
    SetCookie:      true,                  // 设置语言 Cookie
    CookieName:     "lang",
    CookieMaxAge:   86400 * 365,           // 1 年
    CookiePath:     "/",
    CookieSecure:   true,
    CookieHTTPOnly: true,
    CookieSameSite: "Lax",
    Next: func(c fiber.Ctx) bool {        // 跳过中间件
        return c.Path() == "/health"
    },
}

app.Use(i18n.FiberMiddleware(config))
```

### 跳过特定路径

```go
// Fiber
config := i18n.MiddlewareConfig{
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

## 最佳实践

1. **每个应用使用一个 Bundle**：创建一个翻译包并复用
2. **启动时加载翻译**：避免在请求处理时加载
3. **使用基于上下文的翻译**：优先使用 `TFromRequest`/`TFromContext`
4. **设置回退语言**：始终配置缺失翻译的回退
5. **使用有意义的键名**：使用点号分隔，如 `error.not_found`、`page.home.title`

## 许可证

Apache License 2.0
