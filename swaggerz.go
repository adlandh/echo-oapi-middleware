package echooapimiddleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

const defaultPath = "/swaggerz"
const contentTypeYAML = "text/yaml; charset=utf-8"

// SwaggerzConfig configures the swagger YAML endpoint middleware.
type SwaggerzConfig struct {
	// Path is the endpoint path where swagger YAML is served.
	// Default: /swaggerz
	Path string
}

// SwaggerzBytes creates middleware that serves swagger YAML from raw bytes with default config.
func SwaggerzBytes(specBytes []byte) (echo.MiddlewareFunc, error) {
	return SwaggerzBytesWithConfig(specBytes, SwaggerzConfig{})
}

// SwaggerzBytesWithConfig creates middleware that serves swagger YAML from raw bytes.
func SwaggerzBytesWithConfig(specBytes []byte, cfg SwaggerzConfig) (echo.MiddlewareFunc, error) {
	if len(specBytes) == 0 {
		return nil, errors.New("spec bytes must not be empty")
	}

	body := make([]byte, len(specBytes))
	copy(body, specBytes)

	return swaggerzMiddleware(body, cfg), nil
}

// SwaggerzSpec creates middleware that serves swagger YAML from openapi3.T with default config.
func SwaggerzSpec(spec *openapi3.T) (echo.MiddlewareFunc, error) {
	return SwaggerzSpecWithConfig(spec, SwaggerzConfig{})
}

// SwaggerzSpecWithConfig creates middleware that serves swagger YAML from openapi3.T.
func SwaggerzSpecWithConfig(spec *openapi3.T, cfg SwaggerzConfig) (echo.MiddlewareFunc, error) {
	if spec == nil {
		return nil, errors.New("spec must not be nil")
	}

	body, err := yaml.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal openapi spec to yaml: %w", err)
	}

	return swaggerzMiddleware(body, cfg), nil
}

func swaggerzMiddleware(body []byte, cfg SwaggerzConfig) echo.MiddlewareFunc {
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
