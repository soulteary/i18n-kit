# i18n-kit

[![Go Reference](https://pkg.go.dev/badge/github.com/soulteary/i18n-kit/v4.svg)](https://pkg.go.dev/github.com/soulteary/i18n-kit/v4)
[![Go Report Card](.github/goreportcard.svg)](.github/goreportcard-report.md)
[![CI](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml/badge.svg)](https://github.com/soulteary/i18n-kit/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/soulteary/i18n-kit/branch/main/graph/badge.svg)](https://codecov.io/gh/soulteary/i18n-kit)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

一个轻量级、灵活的 Go 国际化 (i18n) 库。支持从 HTTP 请求自动检测语言、翻译包管理，以及 Fiber 和 net/http 双框架中间件。

[English Documentation](README.md)


> **v4.0.0 的破坏性变更 —— 模块路径变了，net/http 与 YAML 各自移进了子包。**
>
> **第一步 —— 所有人，包括完全不提供 HTTP 服务的程序。** 模块路径现在是
> `github.com/soulteary/i18n-kit/v4`：
>
> ```bash
> go get github.com/soulteary/i18n-kit/v4
> go mod edit -droprequire github.com/soulteary/i18n-kit/v3
> ```
>
> 然后改掉源码里的 import 路径。主版本号必须跳，这是 Go 的导入兼容性规则决定的：
> v4 删掉了导出符号。留转发用的空壳不是一个选项 —— 空壳自己就要 import 被搬走的
> 那些包。
>
> **第二步 —— net/http 用户。** 中间件、`*http.Request` 系列 helper 和请求适配器
> 移到了 `github.com/soulteary/i18n-kit/v4/httpadapter`，于是导入根包不再把一个
> 从未启动的 Web 服务器链接进二进制。翻译在打印本地化帮助或错误信息的 CLI 里同样
> 有用，而对只导入根包的程序来说，这意味着
> **少链接 122 个包、二进制小 50%**（203 → 81 个包，3,940,615 → 1,974,432 字节，
> linux/amd64 上 `-trimpath -ldflags="-s -w"` 实测）。
>
> | 改之前 | 改之后 |
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
> `MiddlewareConfig.NextStd` 搬家的理由，和 v3 里 `Next` 搬去 `fiberadapter.Config`
> 是同一个：一个 `func(*http.Request) bool` 字段正是把 net/http 拽进根包的东西。
> `httpadapter.Config` 内嵌 `i18n.MiddlewareConfig` 再加一个 `Next`，与
> `fiberadapter.Config` 完全对称。
>
> **第三步 —— 用 YAML 翻译文件的人。** `LoadYAML` 与 `LoadYAMLFile` 移到了
> `github.com/soulteary/i18n-kit/v4/yamlloader`：把一个 YAML 解析器放在根包里，
> 账是每个程序都要付的，包括翻译文件全是 JSON 的那大多数 —— 而 JSON 标准库本来就读。
>
> | 改之前 | 改之后 |
> |---|---|
> | `bundle.LoadYAML(lang, data)` | `yamlloader.Load(bundle, lang, data)` |
> | `bundle.LoadYAMLFile(lang, path)` | `yamlloader.LoadFile(bundle, lang, path)` |
> | `bundle.LoadDirectory(dir)`（目录里有 .yaml/.yml） | `yamlloader.LoadDirectory(bundle, dir)` |
>
> **`Bundle.LoadDirectory` 现在只读 `.json`。** 目录里出现 `.yaml` 或 `.yml` 时它会
> 返回一个点名 `yamlloader.LoadDirectory` 的错误，而不是默默只加载一半 —— 这是本次
> 唯一的行为变化，而且是刻意吵的。纯 JSON 的目录不受影响。
>
> **两样都不用的话，第二、三步可以跳过。** Bundle、`Translator`、`Detector`、
> `RequestSource`、context helper、`MiddlewareConfig`、`ResolveMiddlewareConfig`、
> `ResolveCookieSameSite`，以及所有格式化与复数函数都留在根包，签名与 v3 一致。
> `fiberadapter` 只有 import 路径要改。
>
> **根包里确实有一个签名变了：** `Formatter.FormatWithContext` 现在接收
> `context.Context`，不再是 `interface{}`。它从来就没能通过旧签名工作过 ——
> 它用的类型分支匹配不到任何类型，于是每一次调用都按 `DefaultLanguage` 格式化
> —— 详见升级说明第 5 步。
>
> → **[升级说明（v4.0.0）](#升级说明v400)**

## 特性

- **多语言支持**：内置支持 10+ 种语言（EN, ZH, FR, DE, JA, KO, IT, ES, PT, RU）
- **语言检测**：自动从查询参数、Cookie、Header 和 Accept-Language 检测语言
- **翻译包**：线程安全的翻译管理，支持回退机制
- **不绑定框架**：net/http 在 `httpadapter`，Fiber v3 在 `fiberadapter`，其他任何框架约 20 行即可接入 —— 检测面向的是一个三方法接口
- **用到才付出代价**：根包不依赖标准库之外的任何东西，连 `net/http` 都不导入 —— Web 服务器、Fiber、YAML 解析器各自待在自己的子包后面
- **上下文集成**：从 context 存取语言信息
- **命名参数**：支持 `{name}` 风格的参数替换
- **复数形式**：简单的复数形式处理
- **文件加载**：JSON 开箱即用；YAML 经由 `yamlloader` 子包

## 环境要求

- **Go 1.27+**（`go.mod` 声明 `go 1.27.0`）
- `github.com/gofiber/fiber/v3` v3.5.0+ —— 只有 `fiberadapter` 子包需要
- `go.yaml.in/yaml/v3` —— 只有 `yamlloader` 子包需要
- `httpadapter` 子包与根包只需要标准库

v4 模块线面向 Fiber v3，且只有 `fiberadapter` 子包会链接它。
仍在 Fiber v2 上的应用请继续使用 `github.com/soulteary/i18n-kit` v1。

## 安装

```bash
go get github.com/soulteary/i18n-kit/v4
```

根包不依赖标准库之外的任何东西，连 `net/http` 都不依赖。需要依赖的部分都在各自的
子包里，二进制只链接它真正用到的那些：

```bash
# net/http 中间件与 *http.Request helper —— 只用标准库
go get github.com/soulteary/i18n-kit/v4/httpadapter

# YAML 翻译文件 —— 会链接 go.yaml.in/yaml/v3
go get github.com/soulteary/i18n-kit/v4/yamlloader

# Fiber v3 中间件 —— 会链接 Fiber，连带 fasthttp
go get github.com/soulteary/i18n-kit/v4/fiberadapter
```

Fiber 集成要求 Fiber v3.5.0 或更高版本 —— 也就是 `go.mod` 里要求的版本。仍使用 Fiber v2 的应用应继续使用 `github.com/soulteary/i18n-kit` v1。

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    i18n "github.com/soulteary/i18n-kit/v4"
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
    i18n "github.com/soulteary/i18n-kit/v4"
    "github.com/soulteary/i18n-kit/v4/httpadapter"
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
        greeting := httpadapter.T(r, "greeting")
        w.Write([]byte(greeting))
    })

    // 应用中间件
    http.Handle("/", httpadapter.Middleware()(handler))
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
    i18n "github.com/soulteary/i18n-kit/v4"
    "github.com/soulteary/i18n-kit/v4/fiberadapter"
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
lang := detector.Detect(httpadapter.RequestSourceOf(r))  // 同一件事的显式写法
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
// JSON —— 只用标准库，不引入额外依赖
bundle.LoadJSONFile(i18n.LangEN, "locales/en.json")
bundle.LoadJSON(i18n.LangEN, data)   // 从字节加载：embed.FS、HTTP 响应体等

// YAML —— 会链接 go.yaml.in/yaml/v3，因此放在 yamlloader 子包里
yamlloader.LoadFile(bundle, i18n.LangZH, "locales/zh.yaml")
yamlloader.Load(bundle, i18n.LangZH, data)

// 加载整个目录，文件命名为 en.json、zh.yaml、fr.json ……
bundle.LoadDirectory("locales/")               // 只读 .json
yamlloader.LoadDirectory(bundle, "locales/")   // .json、.yaml 和 .yml
```

`Bundle.LoadDirectory` **只读 `.json`**。目录里出现 `.yaml` 或 `.yml` 时，它会返回
一个点名 `yamlloader.LoadDirectory` 的错误，而不是默默只加载一半翻译。
`yamlloader.LoadDirectory` 仍然包含 `.json`，所以把其中一个文件改名成 `.yaml`
不会让其余文件停止加载。

#### 让加载器认识其他格式

`Decoder` 负责把单个文件的字节解析成一个扁平 map —— 根包正是靠它加载那些自己
并不链接解析器的格式：

```go
type Decoder func(data []byte) (map[string]string, error)
```

`LoadDirectoryWith` 接收它应当认识的扩展名：

```go
bundle.LoadDirectoryWith("locales/", map[string]i18n.Decoder{
    ".json": i18n.DecodeJSON,   // LoadDirectory 用的就是它
    ".yaml": yamlloader.Decode, // yamlloader.LoadDirectory 额外加的就是它
    ".toml": myTOMLDecoder,     // 其他格式同理，三行就够
})
```

`yamlloader.LoadDirectory` 支持 YAML 而根包始终不 import YAML 解析器，用的正是
这个机制 —— 同一个接缝也对你自己的格式开放。

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

`i18n.TWithLang` / `i18n.TfWithLang` 在每次调用时指定语言，完全不读全局值，这是
规避这种共享最简单的办法：

```go
i18n.TWithLang(i18n.LangZH, "greeting")              // "你好，世界！"
i18n.TfWithLang(i18n.LangZH, "greeting", "Alice")
```

### 作用域内的 Translator

`Translator` 把一个 bundle 和一个当前语言绑定在一起，不与进程中其他部分共享任何状态：

```go
translator := i18n.NewTranslator(bundle)                           // 从 DefaultLanguage 开始
translator = i18n.NewTranslatorWithLanguage(bundle, i18n.LangZH)   // 也可以指定起始语言

translator.SetLanguage(i18n.LangFR)
translator.GetLanguage()                        // "fr"
translator.Bundle()                             // 构造时传入的那个 bundle

translator.T("greeting")                        // 使用 translator 自己的语言
translator.Tf("greeting", "Alice")
translator.TWithLang(i18n.LangZH, "greeting")   // 单次调用覆盖语言
translator.TfWithLang(i18n.LangZH, "greeting", "Alice")
```

`i18n.GlobalTranslator` 就是 `T`/`Tf` 背后那个包级 translator。它是共享值，所以
请优先使用你自己的 translator，或下文的上下文 helper，而不是去修改它。

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

`i18n.Format` 和 `i18n.Pluralize` 只是 `i18n.DefaultFormatter` 的薄封装 —— 后者
是构建在 `DefaultBundle` 之上的 `*Formatter`。只要用的不是进程级的那套翻译，就用
`NewFormatter` 自己造一个。

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
httpadapter.Tf(r, "greeting", "Alice", 30)
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
lang := i18n.LanguageFromContext(ctx)          // 没有时返回 DefaultLanguage
lang, ok := i18n.LanguageFromContextOK(ctx)    // 没有时返回 ("", false)
translation := i18n.TFromContext(ctx, "greeting")
formatted := i18n.TfFromContext(ctx, "greeting", "Alice")
```

当你需要把「这个 context 里没有语言」和「语言就是默认值」区分开时，用
`LanguageFromContextOK` —— `LanguageFromContext` 会把两者混为一谈。

### 在上下文中携带 Bundle

上下文除了语言，还能携带一个 bundle —— 多租户、每个租户一套翻译的服务正需要这个：

```go
ctx = i18n.ContextWithBundle(ctx, tenantBundle)

bundle := i18n.BundleFromContext(ctx)                  // 没有时返回 DefaultBundle
text := i18n.TFromContextWithBundle(ctx, "greeting")
text = i18n.TfFromContextWithBundle(ctx, "greeting", "Alice")
```

带 `WithBundle` 的变体从上下文里取 bundle；而普通的 `TFromContext` /
`TfFromContext` 始终使用 `DefaultBundle`。`httpadapter` 的中间件两个值都会写入，
所以 handler 里的 `r.Context()` 会带上配置指定的那个 bundle。

`Formatter` 同样可以从上下文解析语言：

```go
formatter.FormatWithContext(r.Context(), "welcome", map[string]any{"name": "Alice"})
```

它使用的始终是 formatter **自己的** bundle —— `ContextWithBundle` 不会覆盖它；
上下文里没有语言时回退到 `DefaultLanguage`。Fiber 把语言放在 `Locals` 而不是
context 里：用 `fiberadapter.Language(c)` 取出来再直接调用 `Format`。

### 使用 http.Request

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // 从请求上下文获取语言
    lang := httpadapter.Language(r)
    
    // 获取翻译
    greeting := httpadapter.T(r, "greeting")
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

`ResolveCookieSameSite` 返回一个 `i18n.CookieSameSiteMode` —— `SameSiteLax`、
`SameSiteStrict`、`SameSiteNone`、`SameSiteDisabled` 四者之一；`mode.RequiresSecure()`
则报告适配器必须照办的 `SameSiteNone` 这一种情况。`httpadapter.SameSite(mode)`
负责把它转成 net/http 需要的 `http.SameSite` 常量：

```go
mode := i18n.ResolveCookieSameSite(cfg.CookieSameSite)   // i18n.SameSiteLax
if mode.RequiresSecure() {
    cookie.Secure = true
}
sameSite, write := httpadapter.SameSite(mode)            // "disabled" 时 write 为 false
```

`i18n.DefaultMiddlewareConfig()` 返回这些字段回退到的默认值；而
`i18n.ResolveMiddlewareConfig(cfg...)` 是每个适配器用来处理调用方配置的入口 ——
自己适配框架时请用它，好让默认值只有一处定义。

### 跳过特定路径

```go
// Fiber
config := fiberadapter.Config{
    Next: func(c fiber.Ctx) bool {
        return c.Path() == "/api/internal"
    },
}

// net/http
config := httpadapter.Config{
    Next: func(r *http.Request) bool {
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
| 适配 `*http.Request` | `httpadapter.RequestSourceOf(r)` |
| 就"结果存哪里"达成一致 | `i18n.LocalsLanguageKey`、`i18n.LocalsBundleKey` |

一个完整的适配器，以 Echo 为例：

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

i18n.SupportedLanguages                        // 上表这些内置语言的 []Language
```

`SupportedLanguages` 就是 `IsValid` 用来比对的那个切片，`RegisterLanguage` 会往里
追加 —— 要做语言切换器就遍历它，不必把上面那张表硬编码一遍。

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

## 升级说明（v4.0.1）

API 没有任何变化。`yamlloader` 换了一个模块来解析 YAML，`go.mod` 里少了一个依赖。

- **YAML 解析器改为 `go.yaml.in/yaml/v3` v3.0.5。** `gopkg.in/yaml.v3` 上游已归档
  —— 仓库只读，不会再有任何提交进去，包括修复。`go.yaml.in/yaml/v3` 是它的维护中
  的延续：包名一样，`yaml.Unmarshal` 一样，解析行为一样。代码里唯一改动的一行就是
  `yamlloader/yamlloader.go` 里的 import path。
- **你的代码不需要做任何改动。** `yamlloader` 导出的名字 —— `Decode`、`Load`、
  `LoadFile`、`LoadDirectory` —— 没有一个在签名里出现 YAML 库的类型，它们只处理
  `map[string]string` 和 `i18n` 的类型。升级就是一句
  `go get github.com/soulteary/i18n-kit/v4@v4.0.1`，没有别的。
- **重复的解析器没有了。** v4.0.0 同时要求两个：`gopkg.in/yaml.v3` v3.0.1 是
  `yamlloader` 的直接依赖，`go.yaml.in/yaml/v3` v3.0.5 是间接依赖 —— 因为
  `gofiber/utils/v2` 和 `stretchr/testify` 早就换过去了。同时用到 `yamlloader` 和
  `fiberadapter` 的构建，会链接同一个解析器的两份副本。现在 `gopkg.in/yaml.v3` 在
  `go.mod` 和 `go.sum` 里都不存在了。
- **如果你自己的代码 import 了 `gopkg.in/yaml.v3`，它现在是你的依赖。** v4.0.0 把它
  列为直接依赖，所以它本来就在你的 module graph 里；v4.0.1 不再列它。Fiber v3.5.0
  的 `go.mod` 里仍然 require 它，但构建中没有任何包 import 它，所以它不进 build
  list。这不会悄悄坏掉 —— `go mod tidy` 会把这行加到你的 `go.mod` 里 —— 但这行从
  此归你维护。
- **根包依然完全不链接 YAML 解析器。** 这正是 `yamlloader` 存在的意义；
  `deps_test.go` 里的守卫现在对两个模块路径都会失败，而不只是旧的那个。

## 升级说明（v4.0.0）

这是一份机械式清单；每条背后的原因见本文件顶部的说明。

1. **改 import path** 为 `github.com/soulteary/i18n-kit/v4`，每个文件都要改：

   ```bash
   go get github.com/soulteary/i18n-kit/v4
   go mod edit -droprequire github.com/soulteary/i18n-kit/v3
   ```

   `go get -u` 不会帮你做这件事；v3 停留在 `v3.0.0`。

2. **把 net/http 入口改指向** `github.com/soulteary/i18n-kit/v4/httpadapter`
   —— 顶部说明里的表格列全了这十一个。名字去掉了包名已经表达的那部分：
   `StdMiddleware` 成了 `httpadapter.Middleware`，`LanguageFromRequest` 成了
   `httpadapter.Language`，与 `fiberadapter` 对称。

3. **把 `MiddlewareConfig.NextStd` 移到** `httpadapter.Config` 上，该结构体内嵌
   `i18n.MiddlewareConfig`，与 `fiberadapter.Config` 处理 Fiber 的 `Next` 完全一样：

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

4. **把 YAML 加载改指向** `github.com/soulteary/i18n-kit/v4/yamlloader`：

   ```diff
   -err := bundle.LoadYAMLFile(i18n.LangZH, "locales/zh.yaml")
   +err := yamlloader.LoadFile(bundle, i18n.LangZH, "locales/zh.yaml")
   ```

   **如果你加载的是一个含 YAML 的目录，这一条不会编译报错。**
   `Bundle.LoadDirectory` 仍然存在、仍然能编译；它现在只读 `.json`，遇到 `.yaml`
   或 `.yml` 会返回一个点名 `yamlloader.LoadDirectory` 的错误。之所以让它吵，正是
   因为翻译只加载了一半这件事在有人报告「页面上一堆没翻译的 key」之前是看不见的。
   纯 JSON 的目录不受影响。

   `LoadDirectoryWith` 是新增的，把它认识的扩展名作为 `map[string]i18n.Decoder`
   传进来 —— `yamlloader.LoadDirectory` 就是这样在不让根包 import YAML 库的前提下
   加上 YAML 的。想要 TOML 或 `.properties` 的话，二十行，这边一行都不用改。

5. **`Formatter.FormatWithContext` 现在接收 `context.Context`。** 这一条是修复，
   无论你用不用 net/http 或 YAML 都适用：

   ```diff
   -func (f *Formatter) FormatWithContext(ctx interface{}, key string, params map[string]interface{}) string
   +func (f *Formatter) FormatWithContext(ctx context.Context, key string, params map[string]interface{}) string
   ```

   它过去接收 `interface{}`，并用 `interface{ Context() interface{} }` 这个类型
   分支去识别 Fiber 的 context。**从来没有任何类型满足过这个分支** —— `fiber.Ctx`
   声明的是 `Context() context.Context`，不是 `Context() interface{}` —— 所以每一次
   调用都落到了 `DefaultLanguage`，连传入由 `ContextWithLanguage` 构造的真正
   `context.Context` 也一样。这个方法自己没有测试，这正是它一直没被发现的原因。
   现在它通过 `LanguageFromContext` 解析语言：

   ```go
   // net/http —— 中间件会把语言写进请求的 context
   formatter.FormatWithContext(r.Context(), "welcome", params)

   // Fiber —— 语言在 Locals 里，不在 context 里
   formatter.Format(fiberadapter.Language(c), "welcome", params)
   ```

   原本就传 `context.Context` 的代码照常编译，并且开始真正拿到 context 里的语言，
   而不是永远拿默认值。传其他东西则会编译报错 —— 这正是目的所在，因为它以前只会
   悄悄给出错误答案。

6. **net/http 和 YAML 都不用的话，第 2 到 4 步都不适用。** Bundle、`Translator`、
   `Detector`、`RequestSource`、context helper、`MiddlewareConfig`、
   `ResolveMiddlewareConfig`、`ResolveCookieSameSite`，以及除 `FormatWithContext`
   之外的所有格式化与复数函数，签名都与 v3 一致地留在根包。`fiberadapter` 的用户
   只有 import 路径要改。

除此之外没有任何变化：检测规则、Cookie 属性、翻译输出都一样；格式化行为除上面那处
`FormatWithContext` 的修复外也没有变动。

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
