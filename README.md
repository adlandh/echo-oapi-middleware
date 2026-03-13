# Echo OpenAPI Middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/adlandh/echo-oapi-middleware/v2.svg)](https://pkg.go.dev/github.com/adlandh/echo-oapi-middleware/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/adlandh/echo-oapi-middleware)](https://goreportcard.com/report/github.com/adlandh/echo-oapi-middleware)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue)](https://go.dev/)

Echo middleware for serving OpenAPI/Swagger specifications and UI. Type-safe, zero-dependency (except for Echo and kin-openapi).

## Features

| Feature | Description |
|---------|-------------|
| 🚀 **Type-Safe** | Uses `*openapi3.T` — no runtime parsing errors |
| 📄 **YAML Auto-Serialization** | Specs serialized to YAML automatically |
| 🎨 **Swagger UI** | Built-in UI with CDN (works offline with local assets) |
| ⚙️ **Fully Configurable** | Custom paths for UI and spec endpoints |
| 🛡️ **Non-Mutating** | Original spec object is never modified |
| 📝 **HEAD Support** | Correct Content-Length for HEAD requests |
| ✅ **Well Tested** | 100% test coverage with table-driven tests |

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Advanced Usage](#advanced-usage)
- [Important Notes](#important-notes)
- [Performance](#performance)
- [Requirements](#requirements)
- [License](#license)

## Installation

```bash
go get github.com/adlandh/echo-oapi-middleware/v2
```

## Quick Start

### Minimal Example (YAML Only)

```go
package main

import (
    "github.com/getkin/kin-openapi/openapi3"
    "github.com/labstack/echo/v5"
    echooapimiddleware "github.com/adlandh/echo-oapi-middleware/v2"
)

func main() {
    e := echo.New()

    spec := &openapi3.T{
        OpenAPI: "3.0.3",
        Info:    &openapi3.Info{Title: "My API", Version: "1.0.0"},
        Paths:   &openapi3.Paths{},
    }

    // Serves YAML at GET /swagger.yaml
    e.Use(echooapimiddleware.SwaggerYaml(spec))

    e.GET("/api/users", func(c *echo.Context) error {
        return c.JSON(200, []string{"user1", "user2"})
    })

    e.Start(":8080")
}
```

### With Swagger UI

```go
func main() {
    e := echo.New()

    spec := &openapi3.T{
        OpenAPI: "3.0.3",
        Info:    &openapi3.Info{Title: "Pet Store", Version: "1.0.0"},
        Paths:   &openapi3.Paths{},
    }

    // Serves UI at GET /swagger, /swagger/, /swagger/index.html
    // Serves YAML at GET /swagger.yaml
    e.Use(echooapimiddleware.SwaggerUI(spec))

    e.Start(":8080")
}
```

### With oapi-codegen

If you use [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate your API:

```go
import (
    "github.com/labstack/echo/v5"
    echooapimiddleware "github.com/adlandh/echo-oapi-middleware/v2"
    "your-generated-api/pkg/api"
)

func main() {
    e := echo.New()

    spec, _ := api.GetSwagger() // From generated code

    e.Use(echooapimiddleware.SwaggerUI(spec))
    // ...
}
```

## API Reference

### SwaggerYaml

Serves OpenAPI spec as YAML.

```go
func SwaggerYaml(spec *openapi3.T) echo.MiddlewareFunc
```

| Endpoint | Method | Content-Type |
|----------|--------|--------------|
| `/swagger.yaml` | GET, HEAD | `text/yaml; charset=utf-8` |

### SwaggerYamlWithConfig

Serves OpenAPI spec with custom configuration.

```go
func SwaggerYamlWithConfig(spec *openapi3.T, cfg SwaggerYamlConfig) echo.MiddlewareFunc
```

### SwaggerUI

Serves Swagger UI + YAML spec.

```go
func SwaggerUI(spec *openapi3.T) echo.MiddlewareFunc
```

| Endpoint | Method | Content-Type |
|----------|--------|--------------|
| `/swagger` | GET, HEAD | `text/html; charset=utf-8` |
| `/swagger/` | GET, HEAD | `text/html; charset=utf-8` |
| `/swagger/index.html` | GET, HEAD | `text/html; charset=utf-8` |
| `/swagger.yaml` | GET, HEAD | `text/yaml; charset=utf-8` |

### SwaggerUIWithConfig

Serves Swagger UI with custom configuration.

```go
func SwaggerUIWithConfig(spec *openapi3.T, cfg SwaggerUIConfig) echo.MiddlewareFunc
```

## Configuration

### SwaggerYamlConfig

```go
type SwaggerYamlConfig struct {
    // Path for YAML endpoint.
    // Default: "/swagger.yaml"
    Path string

    // KeepServers preserves the servers field in output.
    // Default: false (servers are stripped for CORS compatibility)
    KeepServers bool
}
```

### SwaggerUIConfig

```go
type SwaggerUIConfig struct {
    // Path for Swagger UI.
    // Default: "/swagger"
    Path string

    // SpecPath for YAML endpoint.
    // Default: "/swagger.yaml"
    SpecPath string

    // KeepServers preserves the servers field in output.
    // Default: false
    KeepServers bool
}
```

## Advanced Usage

### Custom Paths

```go
e.Use(echooapimiddleware.SwaggerUIWithConfig(spec, echooapimiddleware.SwaggerUIConfig{
    Path:     "/api/docs",      // UI at /api/docs
    SpecPath: "/api/openapi.yaml", // YAML at /api/openapi.yaml
}))
```

### Multiple API Versions

```go
func main() {
    e := echo.New()

    // v1 API
    e.Use(echooapimiddleware.SwaggerUIWithConfig(&openapi3.T{
        OpenAPI: "3.0.3",
        Info:    &openapi3.Info{Title: "Pet Store", Version: "1.0.0"},
        Paths:   &openapi3.Paths{},
    }, echooapimiddleware.SwaggerUIConfig{
        Path:     "/v1/docs",
        SpecPath: "/v1/openapi.yaml",
    }))

    // v2 API
    e.Use(echooapimiddleware.SwaggerUIWithConfig(&openapi3.T{
        OpenAPI: "3.0.3",
        Info:    &openapi3.Info{Title: "Pet Store", Version: "2.0.0"},
        Paths:   &openapi3.Paths{},
    }, echooapimiddleware.SwaggerUIConfig{
        Path:     "/v2/docs",
        SpecPath: "/v2/openapi.yaml",
    }))

    e.Start(":8080")
}
```

### Preserve Servers

By default, servers are stripped from output (better CORS). To preserve:

```go
e.Use(echooapimiddleware.SwaggerYamlWithConfig(spec, echooapimiddleware.SwaggerYamlConfig{
    KeepServers: true,
}))
```

### Load Spec from File

```go
import (
    "io"
    "os"
    "gopkg.in/yaml.v3"
)

func main() {
    f, _ := os.Open("openapi.yaml")
    defer f.Close()

    data, _ := io.ReadAll(f)

    var spec openapi3.T
    yaml.Unmarshal(data, &spec)

    e.Use(echooapimiddleware.SwaggerUI(&spec))
}
```

## Important Notes

### Non-Mutating

The middleware never modifies your spec object:

```go
spec := &openapi3.T{
    OpenAPI: "3.0.3",
    Info:    &openapi3.Info{Title: "API"},
    Servers: openapi3.Servers{{URL: "https://api.example.com"}},
}

e.Use(echooapimiddleware.SwaggerYaml(spec))

// spec.Servers is STILL here - not mutated!
_ = spec.Servers // safe
```

### Request Methods

- **GET** — returns full response
- **HEAD** — returns headers only (Content-Length set correctly)
- Other methods pass through to next handler

### CDN Dependencies

Swagger UI loads from unpkg CDN:
- `https://unpkg.com/swagger-ui-dist@5/swagger-ui.css`
- `https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js`

For offline use, you'll need to serve these assets separately or use a local copy.

## Performance

| Metric | Value |
|--------|-------|
| Serialization | Once at middleware creation |
| Request overhead | ~350ns |
| Memory per request | O(1) |

## Requirements

- Go 1.25+
- [Echo v5](https://echo.labstack.com/) — `github.com/labstack/echo/v5`
- [kin-openapi](https://github.com/getkin/kin-openapi) — `github.com/getkin/kin-openapi`

## License

MIT — see [LICENSE](LICENSE) file.
