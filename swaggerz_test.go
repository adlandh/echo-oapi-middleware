package echooapimiddleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

func TestSwaggerzBytes_RequestRouting(t *testing.T) {
	spec := []byte("openapi: 3.0.0\n")

	tests := []struct {
		name          string
		cfg           SwaggerzConfig
		method        string
		path          string
		wantStatus    int
		wantBody      string
		wantType      string
		wantLength    string
		assertSwagger bool
		expectNoBody  bool
	}{
		{
			name:          "get default path",
			method:        http.MethodGet,
			path:          "/swaggerz",
			wantStatus:    http.StatusOK,
			wantBody:      string(spec),
			wantType:      "application/yaml; charset=utf-8",
			assertSwagger: true,
		},
		{
			name:          "head default path",
			method:        http.MethodHead,
			path:          "/swaggerz",
			wantStatus:    http.StatusOK,
			wantType:      "application/yaml; charset=utf-8",
			wantLength:    strconv.Itoa(len(spec)),
			expectNoBody:  true,
			assertSwagger: true,
		},
		{
			name:       "get custom path",
			cfg:        SwaggerzConfig{Path: "/docs/openapi.yaml"},
			method:     http.MethodGet,
			path:       "/docs/openapi.yaml",
			wantStatus: http.StatusOK,
			wantBody:   string(spec),
			wantType:   "application/yaml; charset=utf-8",
		},
		{
			name:       "post swagger path passes through",
			method:     http.MethodPost,
			path:       "/swaggerz",
			wantStatus: http.StatusAccepted,
			wantBody:   "posted",
		},
		{
			name:       "non swagger path passes through",
			method:     http.MethodGet,
			path:       "/users",
			wantStatus: http.StatusOK,
			wantBody:   "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			mw, err := SwaggerzBytesWithConfig(spec, tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			e.Use(mw)
			e.GET("/users", func(c echo.Context) error {
				return c.String(http.StatusOK, "users")
			})
			e.POST("/swaggerz", func(c echo.Context) error {
				return c.String(http.StatusAccepted, "posted")
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantType != "" {
				if got := rec.Header().Get(echo.HeaderContentType); got != tt.wantType {
					t.Fatalf("unexpected content type: %q", got)
				}
			}

			if tt.wantLength != "" {
				if got := rec.Header().Get(echo.HeaderContentLength); got != tt.wantLength {
					t.Fatalf("unexpected content-length: %q", got)
				}
			}

			if tt.expectNoBody {
				if rec.Body.Len() != 0 {
					t.Fatalf("expected empty body, got %q", rec.Body.String())
				}
				return
			}

			if rec.Body.String() != tt.wantBody {
				t.Fatalf("unexpected body: %q", rec.Body.String())
			}
		})
	}
}

func TestSwaggerzSpec_RequestRouting(t *testing.T) {
	tests := []struct {
		name      string
		cfg       SwaggerzConfig
		path      string
		wantParts []string
	}{
		{
			name:      "default path",
			path:      "/swaggerz",
			wantParts: []string{"openapi: 3.1.0", "title: Default", "version: 1.0.0"},
		},
		{
			name:      "custom path",
			cfg:       SwaggerzConfig{Path: "/docs/openapi.yaml"},
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
			mw, err := SwaggerzSpecWithConfig(spec, tt.cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			e.Use(mw)
			e.GET("/users", func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}

			if got := rec.Header().Get(echo.HeaderContentType); got != "application/yaml; charset=utf-8" {
				t.Fatalf("unexpected content type: %q", got)
			}

			if !containsAll(rec.Body.String(), tt.wantParts...) {
				t.Fatalf("unexpected body: %q", rec.Body.String())
			}
		})
	}
}

func TestSwaggerz_BytesInputIsCopied(t *testing.T) {
	e := echo.New()
	spec := []byte("openapi: 3.0.0\n")
	mw, err := SwaggerzBytes(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	spec[0] = 'X'

	e.Use(mw)
	e.GET("/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	req := httptest.NewRequest(http.MethodGet, "/swaggerz", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Body.String() != "openapi: 3.0.0\n" {
		t.Fatalf("unexpected body after input mutation: %q", rec.Body.String())
	}
}

func TestSwaggerz_ConstructorsValidation(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "bytes empty",
			run: func() error {
				_, err := SwaggerzBytes(nil)
				return err
			},
		},
		{
			name: "bytes with config empty",
			run: func() error {
				_, err := SwaggerzBytesWithConfig([]byte{}, SwaggerzConfig{Path: "/docs/openapi.yaml"})
				return err
			},
		},
		{
			name: "spec nil",
			run: func() error {
				_, err := SwaggerzSpec(nil)
				return err
			},
		},
		{
			name: "spec with config nil",
			run: func() error {
				_, err := SwaggerzSpecWithConfig(nil, SwaggerzConfig{Path: "/docs/openapi.yaml"})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err == nil {
				t.Fatal("expected validation error")
			}
		})
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
