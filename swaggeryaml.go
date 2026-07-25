package echooapimiddleware

import (
	"net/http"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
	"gopkg.in/yaml.v3"
)

const (
	defaultSpecPath = "/swagger.yaml"
	contentTypeYAML = "text/yaml; charset=utf-8"
)

// SwaggerYamlConfig configures the swagger YAML endpoint middleware.
type SwaggerYamlConfig struct {
	// Path is the endpoint path where swagger YAML is served.
	// Default: /swagger.yaml
	Path string
	// KeepServers indicates whether to keep the servers field from the spec.
	// Default: false
	KeepServers bool
}

// SwaggerYaml creates middleware that serves swagger YAML from openapi3.T with default config.
func SwaggerYaml(spec *openapi3.T) echo.MiddlewareFunc {
	return SwaggerYamlWithConfig(spec, SwaggerYamlConfig{})
}

// SwaggerYamlWithConfig creates middleware that serves swagger YAML from openapi3.T.
// If spec is nil, the endpoint serves an empty 200 OK. If marshalling fails, the
// endpoint serves 500 at request time so callers can observe the failure.
func SwaggerYamlWithConfig(spec *openapi3.T, cfg SwaggerYamlConfig) echo.MiddlewareFunc {
	if spec == nil {
		return swaggerYamlMiddleware(nil, nil, cfg)
	}

	toMarshal := any(spec)
	if !cfg.KeepServers {
		// Shallow-copy the struct so we can null out Servers without
		// mutating the caller's spec. Nested fields are still shared,
		// but yaml.Marshal only reads them.
		specCopy := *spec
		specCopy.Servers = nil
		toMarshal = &specCopy
	}

	body, marshalErr := yaml.Marshal(toMarshal)

	return swaggerYamlMiddleware(body, marshalErr, cfg)
}

func swaggerYamlMiddleware(body []byte, marshalErr error, cfg SwaggerYamlConfig) echo.MiddlewareFunc {
	path := cfg.Path
	if path == "" {
		path = defaultSpecPath
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()

			if req.URL.Path != path || (req.Method != http.MethodGet && req.Method != http.MethodHead) {
				return next(c)
			}

			if marshalErr != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to marshal openapi spec").Wrap(marshalErr)
			}

			return servePrecomputed(c, contentTypeYAML, body)
		}
	}
}

//nolint:wrapcheck // Framework response helpers return directly.
func servePrecomputed(c *echo.Context, contentType string, body []byte) error {
	c.Response().Header().Set(echo.HeaderContentType, contentType)

	if c.Request().Method == http.MethodHead {
		c.Response().Header().Set(echo.HeaderContentLength, strconv.Itoa(len(body)))
		return c.NoContent(http.StatusOK)
	}

	return c.Blob(http.StatusOK, contentType, body)
}
