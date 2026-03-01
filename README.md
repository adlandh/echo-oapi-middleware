# Echo OpenAPI Middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/adlandh/echo-oapi-middleware.svg)](https://pkg.go.dev/github.com/adlandh/echo-oapi-middleware)
[![Go Report Card](https://goreportcard.com/badge/github.com/adlandh/echo-oapi-middleware)](https://goreportcard.com/report/github.com/adlandh/echo-oapi-middleware)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Echo middleware for serving OpenAPI/Swagger specifications and UI. Provides type-safe handling of OpenAPI specs with automatic YAML serialization and optional Swagger UI integration.

## Features

- 🚀 **Type-Safe API** - Use `*openapi3.T` for compile-time validation
- 📄 **YAML Serving** - Automatically serialize OpenAPI specs to YAML
- 🎨 **Swagger UI** - Optional Swagger UI integration with CDN
- ⚙️ **Configurable Paths** - Customize endpoint paths for specs and UI
- 🛡️ **Non-Mutating** - Specs are never modified by middleware
- 📝 **HEAD Support** - Proper HEAD request handling with Content-Length
- 🎯 **Zero Errors** - Graceful handling of nil/empty inputs
- ✅ **Well-Tested** - Comprehensive test coverage with table-driven tests

## Installation

```bash
go get github.com/adlandh/echo-oapi-middleware
```

## Quick Start

### Basic Usage - Swagger YAML Only

```go
package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	echooapimiddleware "github.com/adlandh/echo-oapi-middleware"
)

func main() {
	e := echo.New()

	// Create your OpenAPI spec
	spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "My API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	// Add middleware to serve the spec at /swagger.yaml
	e.Use(echooapimiddleware.SwaggerYaml(spec))

	// Your routes
	e.GET("/api/users", listUsers)

	e.Start(":8080")
}

func listUsers(c echo.Context) error {
	// Implementation
	return c.JSON(200, []string{})
}
```

### With Swagger UI

```go
package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	echooapimiddleware "github.com/adlandh/echo-oapi-middleware"
)

func main() {
	e := echo.New()

	spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "My API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	// Add middleware to serve both UI and spec
	// UI available at /swagger
	// Spec available at /swagger.yaml
	e.Use(echooapimiddleware.SwaggerUI(spec))

	e.GET("/api/users", listUsers)

	e.Start(":8080")
}

func listUsers(c echo.Context) error {
	return c.JSON(200, []string{})
}
```

## API Reference

### SwaggerYaml

Serves an OpenAPI spec as YAML at a configurable path.

```go
func SwaggerYaml(spec *openapi3.T) echo.MiddlewareFunc
```

**Example:**
```go
e.Use(echooapimiddleware.SwaggerYaml(spec))
// Serves at GET /swagger.yaml
```

### SwaggerYamlWithConfig

Serves an OpenAPI spec with custom configuration.

```go
func SwaggerYamlWithConfig(spec *openapi3.T, cfg SwaggerYamlConfig) echo.MiddlewareFunc
```

**Config Options:**
```go
type SwaggerYamlConfig struct {
	// Path is the endpoint path where swagger YAML is served.
	// Default: /swagger.yaml
	Path string

	// KeepServers indicates whether to keep the servers field from the spec.
	// Default: false (servers field is stripped)
	KeepServers bool
}
```

**Example:**
```go
cfg := echooapimiddleware.SwaggerYamlConfig{
	Path:        "/docs/openapi.yaml",
	KeepServers: true,
}
e.Use(echooapimiddleware.SwaggerYamlWithConfig(spec, cfg))
// Serves at GET /docs/openapi.yaml with servers field preserved
```

### SwaggerUI

Serves both Swagger UI and the OpenAPI spec with default paths.

```go
func SwaggerUI(spec *openapi3.T) echo.MiddlewareFunc
```

**Example:**
```go
e.Use(echooapimiddleware.SwaggerUI(spec))
// UI available at: GET /swagger, /swagger/, /swagger/index.html
// Spec available at: GET /swagger.yaml
```

### SwaggerUIWithConfig

Serves Swagger UI with custom configuration.

```go
func SwaggerUIWithConfig(spec *openapi3.T, cfg SwaggerUIConfig) echo.MiddlewareFunc
```

**Config Options:**
```go
type SwaggerUIConfig struct {
	// Path is the endpoint path where swagger UI is served.
	// Default: /swagger
	Path string

	// SpecPath is the endpoint path where swagger YAML is served.
	// Default: /swagger.yaml
	SpecPath string

	// KeepServers indicates whether to keep the servers field from the spec.
	// Default: false (servers field is stripped)
	KeepServers bool
}
```

**Example:**
```go
cfg := echooapimiddleware.SwaggerUIConfig{
	Path:     "/docs",
	SpecPath: "/docs/openapi.yaml",
}
e.Use(echooapimiddleware.SwaggerUIWithConfig(spec, cfg))
// UI available at: GET /docs, /docs/, /docs/index.html
// Spec available at: GET /docs/openapi.yaml
```

## Advanced Usage

### Multiple API Versions

```go
func main() {
	e := echo.New()

	v1Spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "My API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	v2Spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "My API",
			Version: "2.0.0",
		},
		Paths: &openapi3.Paths{},
	}

	// Serve v1 at /api/v1/docs
	e.Use(echooapimiddleware.SwaggerUIWithConfig(v1Spec, echooapimiddleware.SwaggerUIConfig{
		Path:     "/api/v1/docs",
		SpecPath: "/api/v1/openapi.yaml",
	}))

	// Serve v2 at /api/v2/docs
	e.Use(echooapimiddleware.SwaggerUIWithConfig(v2Spec, echooapimiddleware.SwaggerUIConfig{
		Path:     "/api/v2/docs",
		SpecPath: "/api/v2/openapi.yaml",
	}))

	e.Start(":8080")
}
```

### Preserving Servers Configuration

By default, the `servers` field is stripped from the OpenAPI spec for security reasons (allows the UI to work from any origin). If you need to preserve servers:

```go
cfg := echooapimiddleware.SwaggerYamlConfig{
	KeepServers: true,
}
e.Use(echooapimiddleware.SwaggerYamlWithConfig(spec, cfg))
```

### Handling Nil/Empty Specs

The middleware gracefully handles nil or empty specs without errors:

```go
// This is safe - returns 200 OK with empty body
e.Use(echooapimiddleware.SwaggerYaml(nil))
```

## Important Notes

### Non-Mutating Behavior

The middleware **never modifies** the original spec object passed to it. When `KeepServers: false` is used, an internal copy is created for serialization:

```go
spec := &openapi3.T{
	OpenAPI: "3.0.3",
	Info:    &openapi3.Info{Title: "API", Version: "1.0.0"},
	Servers: openapi3.Servers{{URL: "https://api.example.com"}},
}

// Middleware strips servers from response, but original spec is unchanged
e.Use(echooapimiddleware.SwaggerYamlWithConfig(spec, echooapimiddleware.SwaggerYamlConfig{
	KeepServers: false,
}))

// spec.Servers still contains the original server definitions
fmt.Println(len(spec.Servers)) // Output: 1
```

### Type-Safe OpenAPI Specs

This middleware requires `*openapi3.T` objects from the [kin-openapi](https://github.com/getkin/kin-openapi) library. This ensures:

- Compile-time validation of spec structure
- Type-safe construction and manipulation
- Automatic YAML serialization

To work with raw YAML/JSON:

```go
import "gopkg.in/yaml.v3"

// Parse YAML into spec
var spec openapi3.T
err := yaml.Unmarshal(yamlBytes, &spec)
if err != nil {
	// Handle error
}

e.Use(echooapimiddleware.SwaggerYaml(&spec))
```

### Using oapi-codegen Generated Specs

You can use OpenAPI specs generated by [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) directly with this middleware.

If your generated package exposes `GetSwagger()`, pass the returned `*openapi3.T` into middleware:

```go
import api "your/module/internal/api"

spec, err := api.GetSwagger()
if err != nil {
	// Handle error
}

e.Use(echooapimiddleware.SwaggerUI(spec))
```

This is useful when your OpenAPI document is embedded in generated code and you want to serve docs without separately loading YAML files.

### Request Methods

The middleware handles:
- **GET** - Returns the full response body
- **HEAD** - Returns correct `Content-Length` header without body

Other methods and paths pass through to the next handler.

## Performance

The middleware is efficient and non-blocking:

- Spec serialization happens once during middleware creation
- Request handling is O(1) path comparison
- Minimal memory allocations per request
- ~350ns per request overhead for GET/HEAD operations

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run linter
golangci-lint run
```

## Requirements

- Go 1.25+
- [github.com/labstack/echo/v4](https://github.com/labstack/echo) - v4.0+
- [github.com/getkin/kin-openapi](https://github.com/getkin/kin-openapi) - v0.100+

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## See Also

- [Echo Framework](https://echo.labstack.com/) - Web framework
- [kin-openapi](https://github.com/getkin/kin-openapi) - OpenAPI 3 Go library
- [Swagger UI](https://swagger.io/tools/swagger-ui/) - API documentation UI
