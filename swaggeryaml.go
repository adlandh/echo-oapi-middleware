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

	var err error

	if spec != nil {
		if !cfg.KeepServers {
			spec.Servers = nil
		}

		body, err = yaml.Marshal(spec)
		if err != nil {
			slog.Warn("failed to marshal openapi spec to yaml", "error", err)
		}
	}

	return swaggerYamlMiddleware(body, cfg)
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
