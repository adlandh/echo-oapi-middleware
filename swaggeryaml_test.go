package echooapimiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
)

func TestSwaggerYaml_RequestRouting(t *testing.T) {
	tests := []struct {
		name      string
		cfg       SwaggerYamlConfig
		path      string
		wantParts []string
	}{
		{
			name:      "default path",
			path:      "/swagger.yaml",
			wantParts: []string{"openapi: 3.1.0", "title: Default", "version: 1.0.0"},
		},
		{
			name:      "custom path",
			cfg:       SwaggerYamlConfig{Path: "/docs/openapi.yaml"},
			path:      "/docs/openapi.yaml",
			wantParts: []string{"openapi: 3.0.3", "title: API", "version: 1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &openapi3.T{
				OpenAPI: "3.1.0",
				Info:    &openapi3.Info{Title: "Default", Version: "1.0.0"},
			}
			if tt.cfg.Path != "" {
				spec = &openapi3.T{
					OpenAPI: "3.0.3",
					Info:    &openapi3.Info{Title: "API", Version: "1.0.0"},
				}
			}

			e := echo.New()
			e.Use(SwaggerYamlWithConfig(spec, tt.cfg))
			e.GET("/users", func(c *echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			if got := rec.Header().Get(echo.HeaderContentType); got != "text/yaml; charset=utf-8" {
				t.Fatalf("unexpected content type: %q", got)
			}

			if !containsAll(rec.Body.String(), tt.wantParts...) {
				t.Fatalf("unexpected body: %q", rec.Body.String())
			}
		})
	}
}

func TestSwaggerYaml_KeepServers(t *testing.T) {
	tests := []struct {
		name         string
		cfg          SwaggerYamlConfig
		wantContains bool
	}{
		{
			name:         "default strips servers",
			cfg:          SwaggerYamlConfig{},
			wantContains: false,
		},
		{
			name: "keep servers enabled",
			cfg: SwaggerYamlConfig{
				KeepServers: true,
			},
			wantContains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			spec := &openapi3.T{
				OpenAPI: "3.0.3",
				Info: &openapi3.Info{
					Title:   "API",
					Version: "1.0.0",
				},
				Servers: openapi3.Servers{{
					URL: "https://api.example.com",
				}},
			}

			e.Use(SwaggerYamlWithConfig(spec, tt.cfg))
			e.GET("/users", func(c *echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			got := rec.Body.String()
			hasServers := strings.Contains(got, "servers:") && strings.Contains(got, "https://api.example.com")
			if hasServers != tt.wantContains {
				t.Fatalf("servers presence mismatch, want=%v body=%q", tt.wantContains, got)
			}
		})
	}
}

func TestSwaggerYaml_ConstructorsAcceptEmptyInputs(t *testing.T) {
	e := echo.New()
	e.Use(SwaggerYaml(nil))

	req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get(echo.HeaderContentType); got != contentTypeYAML {
		t.Fatalf("unexpected content type: %q", got)
	}

	if rec.Body.String() != "" {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}

	reqHead := httptest.NewRequest(http.MethodHead, "/swagger.yaml", nil)
	recHead := httptest.NewRecorder()
	e.ServeHTTP(recHead, reqHead)

	if recHead.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recHead.Code)
	}

	if got := recHead.Header().Get(echo.HeaderContentLength); got != "0" {
		t.Fatalf("unexpected content-length: %q", got)
	}
}

func TestSwaggerYaml_NotMutatedByMiddleware(t *testing.T) {
	spec := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "API",
			Version: "1.0.0",
		},
		Servers: openapi3.Servers{{
			URL: "https://api.example.com",
		}},
	}

	// Keep a reference to the original servers slice
	originalServersLength := len(spec.Servers)

	// Create middleware with KeepServers=false (servers should be excluded from output)
	e := echo.New()
	e.Use(SwaggerYamlWithConfig(spec, SwaggerYamlConfig{KeepServers: false}))
	e.GET("/users", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify the response does not contain servers
	got := rec.Body.String()
	if strings.Contains(got, "servers:") || strings.Contains(got, "https://api.example.com") {
		t.Fatalf("expected servers to be excluded from response, but found them in: %q", got)
	}

	// CRITICAL: Verify the original spec object was NOT mutated
	if len(spec.Servers) != originalServersLength {
		t.Fatalf("spec.Servers was mutated! original length=%d, current length=%d", originalServersLength, len(spec.Servers))
	}

	if spec.Servers == nil {
		t.Fatal("spec.Servers was set to nil - the spec object was mutated by the middleware")
	}

	if len(spec.Servers) > 0 && spec.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("spec.Servers[0].URL was mutated! expected 'https://api.example.com', got %q", spec.Servers[0].URL)
	}

	// Test with KeepServers=true to ensure it works correctly
	e2 := echo.New()
	e2.Use(SwaggerYamlWithConfig(spec, SwaggerYamlConfig{KeepServers: true}))
	e2.GET("/users", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req2 := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
	rec2 := httptest.NewRecorder()
	e2.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec2.Code)
	}

	// With KeepServers=true, servers should be in the output
	got2 := rec2.Body.String()
	if !strings.Contains(got2, "servers:") || !strings.Contains(got2, "https://api.example.com") {
		t.Fatalf("expected servers to be included in response with KeepServers=true, but not found in: %q", got2)
	}

	// Verify spec still has servers after second middleware call
	if len(spec.Servers) != originalServersLength {
		t.Fatalf("spec.Servers length changed after second middleware call! original length=%d, current length=%d", originalServersLength, len(spec.Servers))
	}
}

func containsAll(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
