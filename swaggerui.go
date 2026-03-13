package echooapimiddleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
)

const defaultUIPath = "/swagger"
const contentTypeHTML = "text/html; charset=utf-8"

// SwaggerUIConfig configures swagger UI middleware paths.
type SwaggerUIConfig struct {
	// Path is the endpoint path where swagger UI is served.
	// Default: /swagger
	Path string

	// SpecPath is the endpoint path where swagger YAML is served.
	// Default: /swagger.yaml
	SpecPath string

	// KeepServers indicates whether to keep the servers field from the spec.
	// Default: false
	KeepServers bool
}

// SwaggerUI creates middleware that serves swagger UI and YAML from openapi3.T.
func SwaggerUI(spec *openapi3.T) echo.MiddlewareFunc {
	return SwaggerUIWithConfig(spec, SwaggerUIConfig{})
}

// SwaggerUIWithConfig creates middleware that serves swagger UI and YAML from openapi3.T.
func SwaggerUIWithConfig(spec *openapi3.T, cfg SwaggerUIConfig) echo.MiddlewareFunc {
	specPath := cfg.SpecPath
	if specPath == "" {
		specPath = defaultPath
	}

	specMW := SwaggerYamlWithConfig(spec, SwaggerYamlConfig{Path: specPath, KeepServers: cfg.KeepServers})

	return swaggerUIMiddleware(specMW, cfg.Path, specPath)
}

func swaggerUIMiddleware(specMW echo.MiddlewareFunc, uiPath, specPath string) echo.MiddlewareFunc {
	if uiPath == "" {
		uiPath = defaultUIPath
	}

	body := []byte(swaggerUIHTML(specPath))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		handler := specMW(next)

		return func(c *echo.Context) error {
			req := c.Request()

			if isSwaggerUIPath(req.URL.Path, uiPath) && (req.Method == http.MethodGet || req.Method == http.MethodHead) {
				c.Response().Header().Set(echo.HeaderContentType, contentTypeHTML)

				if req.Method == http.MethodHead {
					c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(body)))

					return c.NoContent(http.StatusOK)
				}

				return c.Blob(http.StatusOK, contentTypeHTML, body)
			}

			return handler(c)
		}
	}
}

func isSwaggerUIPath(path, uiPath string) bool {
	if path == uiPath || path == uiPath+"/" {
		return true
	}

	return path == uiPath+"/index.html"
}

func swaggerUIHTML(specPath string) string {
	quotedSpecPath := strconv.Quote(specPath)

	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: ` + quotedSpecPath + `,
      dom_id: '#swagger-ui',
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset,
      ],
      layout: "BaseLayout",
    });
  </script>
</body>
</html>
`
}
