# Conform - Declarative Configuration & Validation for Go

<div align="center">

![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-green.svg?style=for-the-badge)
![GitHub](https://img.shields.io/github/stars/alicanli1995/conform?style=for-the-badge)
![CI/CD](https://img.shields.io/github/actions/workflow/status/alicanli1995/conform/test.yml?style=for-the-badge&label=CI/CD)
[![Go Reference](https://img.shields.io/badge/go-reference-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://pkg.go.dev/github.com/alicanli1995/conform)

**The Pydantic for Go** - Type-safe configuration loading, validation, and management in one elegant package.

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Examples](#-examples)

</div>

---

## 🎯 Why Conform?

**Stop juggling multiple libraries.** Conform unifies configuration loading, type conversion, and validation into a single, declarative interface.

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

---

## ✨ Features

### 🎯 Core Features

- **🏷️ Declarative Configuration** - Everything in struct tags, zero boilerplate
- **🚀 Type-Safe Generics** - Full type safety with Go 1.21+ generics
- **🔄 Multi-Source Support** - Environment variables, files (YAML/JSON/TOML), custom sources
- **✅ Built-in Validation** - 20+ validators out of the box (email, URL, IP, numeric ranges, regex, etc.)
- **🔧 Smart Type Coercion** - Automatic conversion for `bool`, `time.Duration`, `[]int`, `time.Time`, `map[string]int`
- **📦 Nested Structs** - Full support with automatic prefix handling
- **🔥 Hot Reload** - Watch for changes and reload automatically
- **💬 Beautiful Errors** - Detailed error messages with field paths, suggestions, and locations
- **🌍 Environment-Specific** - Load different configs for dev/staging/prod
- **🔐 Variable Substitution** - `${VAR_NAME:-default}` syntax in config values and file paths

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

```bash
go get github.com/alicanli1995/conform
```

---

## 🚀 Quick Start

### Basic Example

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

## 🔧 Advanced Usage

### Multiple File Sources

```go
cfg, err := conform.LoadGeneric[Config](
    conform.FromFile("config.yaml"),
    conform.FromFile("secrets.yaml"),
    conform.FromEnv(),
)
```

### Custom Sources

```go
type CustomSource struct{}

func (c *CustomSource) Get(key string) (string, bool) {
    // Your implementation
    return value, found
}

cfg, err := conform.LoadGeneric[Config](
    conform.WithSource(&CustomSource{}),
)
```

### Slices and Arrays

```go
type Config struct {
    Hosts []string `conform:"env=HOSTS,separator=|"`
    Ports []int    `conform:"env=PORTS,separator=,"`
}

// HOSTS=host1|host2|host3
// PORTS=8080,8081,8082
```

### Maps

```go
type Config struct {
    // String "key1=value1,key2=value2" → map[string]string
    StringMap map[string]string `conform:"env=STRING_MAP"`
    
    // String "1=one,2=two" → map[int]string
    IntMap map[int]string `conform:"env=INT_MAP"`
}
```

### Time Parsing

```go
type Config struct {
    StartTime time.Time `conform:"env=START_TIME,format=2006-01-02T15:04:05Z"`
    BirthDate time.Time `conform:"env=BIRTH_DATE,format=2006-01-02"`
}

// START_TIME=2024-01-01T00:00:00Z
// BIRTH_DATE=1990-05-15
```

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

- **[Basic Usage](./examples/basic/main.go)** - Getting started
- **[Generic API](./examples/generic/main.go)** - Type-safe loading
- **[File Configuration](./examples/fileconfig/main.go)** - YAML, JSON, TOML examples
- **[Environment-Specific](./examples/environment/main.go)** - Dev/staging/prod configs
- **[Advanced Features](./examples/advanced/main.go)** - Complex scenarios
- **[Custom Converters & Validators](./examples/custom/main.go)** - Extending Conform
- **[Hot Reload](./examples/hotreload/main.go)** - Dynamic configuration
- **[Beautiful Errors](./examples/errors/main.go)** - Error handling
- **[Real-World Config](./examples/realworld/main.go)** - Production-ready example

---

## 🏗️ Architecture

Conform consists of four main components:

1. **Parser** - Parses struct tags and extracts configuration metadata
2. **Converter** - Converts string values to typed Go values
3. **Validator** - Validates values against declarative rules
4. **Loader** - Orchestrates loading from multiple sources with priority

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with ❤️ for the Go community**

[⭐ Star us on GitHub](https://github.com/alicanli1995/conform) • [📦 pkg.go.dev](https://pkg.go.dev/github.com/alicanli1995/conform) • [📖 Documentation](#-documentation) • [💬 Issues](https://github.com/alicanli1995/conform/issues)

</div>
