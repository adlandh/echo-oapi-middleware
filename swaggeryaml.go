package echooapimiddleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

const defaultPath = "/swagger.yaml"
const contentTypeYAML = "text/yaml; charset=utf-8"

// SwaggerYamlConfig configures the swagger YAML endpoint middleware.
type SwaggerYamlConfig struct {
	// Path is the endpoint path where swagger YAML is served.
	// Default: /swagger.yaml
	Path string
	// KeepServers indicates whether to keep the servers field from the spec.
	// Default: false
	KeepServers bool
}

// SwaggerYamlBytes creates middleware that serves swagger YAML from raw bytes with default config.
func SwaggerYamlBytes(specBytes []byte) echo.MiddlewareFunc {
	return SwaggerYamlBytesWithConfig(specBytes, SwaggerYamlConfig{})
}

// SwaggerYamlBytesWithConfig creates middleware that serves swagger YAML from raw bytes.
func SwaggerYamlBytesWithConfig(specBytes []byte, cfg SwaggerYamlConfig) echo.MiddlewareFunc {
	var body []byte

	if len(specBytes) != 0 {
		body = make([]byte, len(specBytes))
		copy(body, specBytes)
	}

	return swaggerYamlMiddleware(body, cfg)
}

// SwaggerYamlSpec creates middleware that serves swagger YAML from openapi3.T with default config.
func SwaggerYamlSpec(spec *openapi3.T) echo.MiddlewareFunc {
	return SwaggerYamlSpecWithConfig(spec, SwaggerYamlConfig{})
}

// SwaggerYamlSpecWithConfig creates middleware that serves swagger YAML from openapi3.T.
func SwaggerYamlSpecWithConfig(spec *openapi3.T, cfg SwaggerYamlConfig) echo.MiddlewareFunc {
	var body []byte

	if spec == nil {
		return swaggerYamlMiddleware(body, cfg)
	}

	// Create a wrapper that marshals without servers if KeepServers is false.
	// This avoids mutating the caller's spec object.
	wrapper := &specWrapper{spec: spec, keepServers: cfg.KeepServers}

	body, err := yaml.Marshal(wrapper)
	if err != nil {
		slog.Warn("failed to marshal openapi spec to yaml", "error", err)
	}

	return swaggerYamlMiddleware(body, cfg)
}

// specWrapper wraps an openapi3.T and provides custom marshalling that optionally excludes servers.
type specWrapper struct {
	spec        *openapi3.T
	keepServers bool
}

// MarshalYAML implements yaml.Marshaler to marshal the spec without servers if keepServers is false.
func (sw *specWrapper) MarshalYAML() (any, error) {
	if sw.keepServers {
		return sw.spec, nil
	}

	// Marshal to intermediate format, strip servers, then return for marshalling.
	// We need to create a modified version without mutating the original.
	// The simplest approach: marshal the spec, unmarshal into a map, remove servers, return the map.
	marshalled, err := yaml.Marshal(sw.spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openapi spec: %w", err)
	}

	var data map[string]any
	if err := yaml.Unmarshal(marshalled, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal openapi spec: %w", err)
	}

	delete(data, "servers")

	return data, nil
}

func swaggerYamlMiddleware(body []byte, cfg SwaggerYamlConfig) echo.MiddlewareFunc {
	path := cfg.Path
	if path == "" {
		path = defaultPath
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			if req.URL.Path == path && (req.Method == http.MethodGet || req.Method == http.MethodHead) {
				c.Response().Header().Set(echo.HeaderContentType, contentTypeYAML)

				if req.Method == http.MethodHead {
					c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(body)))

					return c.NoContent(http.StatusOK)
				}

				return c.Blob(http.StatusOK, contentTypeYAML, body)
			}

			return next(c)
		}
	}
}
