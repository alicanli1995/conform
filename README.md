# Conform

<div align="center">

![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-green.svg?style=for-the-badge)
![GitHub](https://img.shields.io/github/stars/alicanli1995/conform?style=for-the-badge)
![CI/CD](https://img.shields.io/github/actions/workflow/status/alicanli1995/conform/test.yml?style=for-the-badge&label=CI/CD)
[![Go Reference](https://img.shields.io/badge/go-reference-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://pkg.go.dev/github.com/alicanli1995/conform)

### The Pydantic for Go

**Type-safe configuration loading, validation, and management in one elegant package.**

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Examples](#-examples)

[![Go Report Card](https://goreportcard.com/badge/github.com/alicanli1995/conform)](https://goreportcard.com/report/github.com/alicanli1995/conform)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/alicanli1995/conform)

</div>

---

## 🎯 Why Conform?

**Stop juggling multiple libraries.** Conform unifies configuration loading, type conversion, and validation into a single, declarative interface.

### Key Benefits

- ⚡ **Zero Boilerplate** - Declare everything in struct tags, no manual parsing or validation code
- 🔒 **Type Safety** - Full generics support ensures compile-time type checking
- 🚀 **Production Ready** - Built-in support for environment-specific configs, hot reload, and variable substitution
- 💡 **Developer Experience** - Beautiful, actionable error messages that tell you exactly what's wrong
- 🎨 **Flexible** - Support for multiple sources, custom validators, and converters
- 📦 **Lightweight** - Minimal dependencies, fast performance

### Before Conform ❌

```go
// Load config
viper.SetConfigFile("config.yaml")
viper.ReadInConfig()

// Unmarshal
var cfg Config
viper.Unmarshal(&cfg)

// Validate
validate := validator.New()
if err := validate.Struct(cfg); err != nil {
    // Parse errors...
}

// Type conversion? Manual!
port, _ := strconv.Atoi(viper.GetString("port"))
timeout, _ := time.ParseDuration(viper.GetString("timeout"))
```

### With Conform ✅

```go
type Config struct {
    Port     int           `conform:"env=PORT,default=8080,validate=gte:1024"`
    Timeout  time.Duration `conform:"env=TIMEOUT,default=30s"`
    Database string        `conform:"env=DB_URL,required,validate=url"`
}

cfg, err := conform.LoadGeneric[Config](conform.FromEnv())
// ✨ Done! Type-safe, validated, ready to use.
```

**One struct tag. One function call. Zero boilerplate.**

### Perfect For

- 🏢 **Microservices** - Type-safe configuration across services
- 🚀 **Cloud-Native Apps** - Environment-specific configs for Kubernetes, Docker
- 🔧 **CLI Tools** - Easy configuration management
- 📊 **APIs & Web Services** - Fast, validated config loading
- 🧪 **Testing** - Mock-friendly configuration loading

---

## ✨ Features

### 🎯 Core Features

<div align="center">

| Feature | Description |
|---------|-------------|
| 🏷️ **Declarative Configuration** | Everything in struct tags, zero boilerplate |
| 🚀 **Type-Safe Generics** | Full type safety with Go 1.21+ generics |
| 🔄 **Multi-Source Support** | Environment variables, files (YAML/JSON/TOML), custom sources |
| ✅ **Built-in Validation** | 20+ validators out of the box |
| 🔧 **Smart Type Coercion** | Automatic conversion for complex types |
| 📦 **Nested Structs** | Full support with automatic prefix handling |
| 🔥 **Hot Reload** | Watch for changes and reload automatically |
| 💬 **Beautiful Errors** | Detailed error messages with suggestions |
| 🌍 **Environment-Specific** | Load different configs for dev/staging/prod |
| 🔐 **Variable Substitution** | `${VAR_NAME:-default}` syntax support |

</div>

### 🎨 What Makes Conform Different?

| Feature | Conform | Viper | envconfig | koanf |
|---------|---------|-------|-----------|-------|
| **Type Safety** | ✅ Generics | ❌ | ❌ | ❌ |
| **Validation** | ✅ Built-in | ❌ Requires validator | ❌ | ⚠️ Basic |
| **Error Messages** | ✅ Beautiful | ❌ | ❌ | ⚠️ Basic |
| **Hot Reload** | ✅ Built-in | ✅ WatchConfig | ❌ | ✅ |
| **Environment-Specific** | ✅ Built-in | ⚠️ Manual | ❌ | ✅ |
| **Variable Substitution** | ✅ Built-in | ❌ | ❌ | ✅ |
| **Declarative** | ✅ 100% | ⚠️ Partial | ✅ | ✅ |
| **Zero Boilerplate** | ✅ | ❌ | ⚠️ | ⚠️ |

**Note:** 
- **Viper** requires separate validation library (e.g., `go-playground/validator`)
- **envconfig** is minimal and focused only on environment variables
- **koanf** is a modern alternative but lacks generics and built-in validation

---

## 📦 Installation

### Go Module

```bash
go get github.com/alicanli1995/conform
```

### Import

```go
import "github.com/alicanli1995/conform"
```

### Requirements

- Go 1.21 or higher
- No external dependencies required (except for file format support: YAML, TOML)

---

## 🚀 Quick Start

### Basic Example

Define your configuration struct with `conform` tags:

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/alicanli1995/conform"
)

type Config struct {
    Port     int    `conform:"env=APP_PORT,default=8080,validate=gte:1024"`
    Host     string `conform:"env=APP_HOST,default=localhost,validate=hostname"`
    Database string `conform:"env=DATABASE_URL,required,validate=url"`
}

func main() {
    os.Setenv("APP_PORT", "3000")
    os.Setenv("DATABASE_URL", "postgres://localhost/mydb")
    
    cfg, err := conform.LoadGeneric[Config](conform.FromEnv())
    if err != nil {
        panic(err) // Beautiful error messages!
    }
    
    fmt.Printf("Server: %s:%d\n", cfg.Host, cfg.Port)
}
```

**That's it!** Your configuration is loaded, validated, and ready to use. 🎉

### What Just Happened?

1. ✅ **Loaded** from environment variables
2. ✅ **Converted** types automatically (`string` → `int`)
3. ✅ **Validated** against rules (`gte:1024`, `hostname`, `url`)
4. ✅ **Applied** defaults where needed
5. ✅ **Returned** type-safe config struct

### Try It Yourself

```bash
# Run the example
cd examples/basic && go run main.go
```

---

## 📖 Documentation

### Multi-Source Support

Load from multiple sources with automatic priority (first source wins):

```go
cfg, err := conform.LoadGeneric[Config](
    conform.FromEnv(),                    // Highest priority
    conform.FromFile("secrets.json"),     // Second priority
    conform.FromFile("config.yaml"),      // Third priority
    conform.WithSource(&CustomSource{}),  // Custom source
)
// Priority: env > custom sources > file sources > defaults
```

### Environment-Specific Configuration

Perfect for dev/staging/production environments:

```go
type Config struct {
    Database struct {
        Host string `conform:"file=database.host,default=localhost"`
        Port int    `conform:"file=database.port,default=5432"`
    }
}

// Development
devCfg, _ := conform.LoadGeneric[Config](
    conform.WithEnvironment("development"),
    conform.FromFile("config.${ENV}.yaml"), // Loads config.development.yaml
)

// Production
prodCfg, _ := conform.LoadGeneric[Config](
    conform.WithEnvironment("production"),
    conform.FromFile("config.${ENV}.yaml"), // Loads config.production.yaml
)
```

### Variable Substitution

Use `${VAR_NAME:-default}` syntax in config values:

```go
type Config struct {
    DatabaseURL string `conform:"env=DB_URL,default=postgres://${DB_USER:-postgres}:${DB_PASSWORD}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-mydb}"`
    APIURL      string `conform:"env=API_URL,default=https://api.${ENV:-dev}.example.com"`
}
```

### Smart Type Coercion

Automatic conversion for complex types:

```go
type Config struct {
    // String "true" → bool true
    Debug bool `conform:"env=DEBUG"`
    
    // String "30s" → time.Duration
    Timeout time.Duration `conform:"env=TIMEOUT"`
    
    // String "1,2,3" → []int{1,2,3}
    IDs []int `conform:"env=IDS,separator=,"`
    
    // String "1=one,2=two" → map[int]string{1:"one", 2:"two"}
    Mapping map[int]string `conform:"env=MAPPING"`
    
    // String "2024-01-01" → time.Time
    StartDate time.Time `conform:"env=START,format=2006-01-02"`
}
```

### Beautiful Error Messages

Get detailed, actionable error messages:

```go
cfg, err := conform.LoadGeneric[Config](conform.FromEnv())
if err != nil {
    fmt.Println(err)
    // Output:
    // ❌ Configuration validation failed:
    //
    // 1. Port (APP_PORT): value 80 is too small
    //    Got: 80
    //    Location: env var APP_PORT
    //    💡 Suggestion: Use a value >= 1024 (e.g. 8080)
    //
    // 2. Database.URL (DB_URL): invalid URL format
    //    Got: "not-a-url"
    //    Expected: valid URL with scheme
    //    💡 Suggestion: Format should be: https://example.com
}
```

### Hot Reload

Watch for configuration changes automatically:

```go
watcher, err := conform.Watch[Config](func(newCfg Config) {
    log.Printf("Config reloaded: %+v", newCfg)
    // Update your application state here
}, conform.FromEnv(), conform.FromFile("config.yaml"))

// Thread-safe access
cfg := watcher.Get()

// Stop watching
defer watcher.Stop()
```

### Custom Validators

Register your own validation rules:

```go
conform.RegisterValidator("strong_password", func(val interface{}, params []string) error {
    str := val.(string)
    if len(str) < 12 {
        return fmt.Errorf("password must be at least 12 characters")
    }
    if !hasSpecialChar(str) {
        return fmt.Errorf("password must contain special character")
    }
    return nil
})

type Config struct {
    Password string `conform:"env=PASSWORD,validate=strong_password"`
}
```

### Custom Converters

Convert to custom types:

```go
type CustomType string

conform.RegisterConverter(
    reflect.TypeOf(CustomType("")),
    func(s string) (interface{}, error) {
        return CustomType("custom_" + s), nil
    },
)

type Config struct {
    Custom CustomType `conform:"env=CUSTOM"`
}
```

### Nested Configuration

Full support for nested structs with automatic prefix handling:

```go
type DatabaseConfig struct {
    Host string `conform:"env=HOST,default=localhost"`
    Port int    `conform:"env=PORT,default=5432"`
}

type AppConfig struct {
    Name     string         `conform:"env=APP_NAME,default=MyApp"`
    Database DatabaseConfig `conform:"prefix=DB_"`
}

// Environment variables:
// DB_HOST=db.example.com
// DB_PORT=5432
```

### File Configuration

Support for YAML, JSON, and TOML:

```go
type Config struct {
    Server struct {
        Host string `conform:"file=server.host,default=localhost"`
        Port int    `conform:"file=server.port,default=8080"`
    }
}

// YAML
cfg, _ := conform.LoadGeneric[Config](conform.FromFile("config.yaml"))

// TOML
cfg, _ := conform.LoadGeneric[Config](conform.FromFile("config.toml"))

// JSON
cfg, _ := conform.LoadGeneric[Config](conform.FromFile("config.json"))
```

---

## 📚 Tag Reference

### Source Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `env=VAR_NAME` | Load from environment variable | `env=APP_PORT` |
| `file=key.path` | Load from config file (dot notation) | `file=database.host` |
| `default=value` | Default value if not found | `default=8080` |
| `required` | Field is required (error if missing) | `required` |
| `prefix=PREFIX_` | Prefix for nested structs | `prefix=DB_` |

### Type Conversion Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `format=layout` | Format for time.Time | `format=2006-01-02` |
| `separator=,` | Separator for slices | `separator=\|` |

### Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `validate=rule:param` | Validation rules | `validate=gte:1024,lte:65535` |

---

## ✅ Built-in Validators

### Numeric Validators

- `min:value` - Minimum value/length
- `max:value` - Maximum value/length
- `gte:value` - Greater than or equal
- `lte:value` - Less than or equal
- `eq:value` - Equal to
- `ne:value` - Not equal to

### String Validators

- `email` - Valid email address
- `url` - Valid URL (use `url:https` for HTTPS only)
- `ip` - Valid IP address (IPv4 or IPv6)
- `hostname` - Valid hostname
- `port` - Valid port number (1-65535)
- `alphanum` - Only letters and digits
- `alpha` - Only letters
- `numeric` - Only digits
- `regex:pattern` - Match regex pattern
- `oneof:val1:val2` - One of the specified values
- `len:length` - Exact length

### Password Validators

- `has_upper` - Contains uppercase letter
- `has_lower` - Contains lowercase letter
- `has_digit` - Contains digit
- `has_special` - Contains special character

### General

- `required` - Field is required

---

## 🔧 CLI Tool

> **Note:** CLI tool is currently in development. For now, use the programmatic API for validation.

Validate configuration files programmatically:

```go
cfg, err := conform.LoadGeneric[Config](
    conform.FromFile("config.yaml"),
    conform.FromEnv(),
)
if err != nil {
    fmt.Println(err) // Beautiful error messages
}
```

---

## 📚 Examples

Comprehensive examples available in the [`examples`](./examples) directory:

| Example | Description | Link |
|---------|-------------|------|
| 🚀 **Basic Usage** | Getting started with Conform | [View](./examples/basic/main.go) |
| 🎯 **Generic API** | Type-safe loading with generics | [View](./examples/generic/main.go) |
| 📄 **File Configuration** | YAML, JSON, TOML examples | [View](./examples/fileconfig/main.go) |
| 🌍 **Environment-Specific** | Dev/staging/prod configs | [View](./examples/environment/main.go) |
| ⚡ **Advanced Features** | Complex scenarios | [View](./examples/advanced/main.go) |
| 🔧 **Custom Extensions** | Custom converters & validators | [View](./examples/custom/main.go) |
| 🔥 **Hot Reload** | Dynamic configuration | [View](./examples/hotreload/main.go) |
| 💬 **Error Handling** | Beautiful error messages | [View](./examples/errors/main.go) |
| 🏢 **Real-World** | Production-ready example | [View](./examples/realworld/main.go) |

### Run All Examples

```bash
make run-examples
```

---

## 🗺️ Roadmap

### Upcoming Features

- 🔐 **Secret Management** - HashiCorp Vault, AWS Secrets Manager, Azure Key Vault integration
- 🔄 **Remote Configuration** - etcd integration for distributed config management
- 📊 **JSON Schema Validation** - External schema validation support
- 📝 **Config Documentation** - Auto-generate config docs from struct definitions
- 🔍 **Config Diff/Compare** - Track and compare configuration changes
- 📈 **Metrics & Observability** - Prometheus metrics for config usage monitoring
---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for the Go community**

[⭐ Star us on GitHub](https://github.com/alicanli1995/conform) • [📦 pkg.go.dev](https://pkg.go.dev/github.com/alicanli1995/conform) • [📖 Documentation](#-documentation) • [💬 Issues](https://github.com/alicanli1995/conform/issues) • [🐛 Report Bug](https://github.com/alicanli1995/conform/issues/new)

</div>
