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

func TestSwaggerUIBytes_DefaultPaths(t *testing.T) {
	e := echo.New()
	spec := []byte("openapi: 3.0.0\n")

	mw := SwaggerUIBytes(spec)

	e.Use(mw)
	e.GET("/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	t.Run("ui html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if got := rec.Header().Get(echo.HeaderContentType); got != contentTypeHTML {
			t.Fatalf("unexpected content type: %q", got)
		}

		if !strings.Contains(rec.Body.String(), `url: "/swagger.yaml"`) {
			t.Fatalf("swagger ui does not point to default spec path, body=%q", rec.Body.String())
		}
	})

	t.Run("spec yaml", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if got := rec.Header().Get(echo.HeaderContentType); got != contentTypeYAML {
			t.Fatalf("unexpected content type: %q", got)
		}

		if rec.Body.String() != string(spec) {
			t.Fatalf("unexpected body: %q", rec.Body.String())
		}
	})
}

func TestSwaggerUIBytes_CustomPaths(t *testing.T) {
	e := echo.New()
	mw := SwaggerUIBytesWithConfig([]byte("openapi: 3.0.0\n"), SwaggerUIConfig{
		Path:     "/docs",
		SpecPath: "/docs/openapi.yaml",
	})

	e.Use(mw)
	e.GET("/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	for _, path := range []string{"/docs", "/docs/", "/docs/index.html"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("path %q: expected status %d, got %d", path, http.StatusOK, rec.Code)
		}

		if !strings.Contains(rec.Body.String(), `url: "/docs/openapi.yaml"`) {
			t.Fatalf("path %q: unexpected html body: %q", path, rec.Body.String())
		}
	}
}

func TestSwaggerUISpec_DefaultPaths(t *testing.T) {
	e := echo.New()
	mw := SwaggerUISpec(&openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "API", Version: "1.0.0"},
	})

	e.Use(mw)
	e.GET("/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	body := rec.Body.String()
	if !containsAll(body, "openapi: 3.0.3", "title: API", "version: 1.0.0") {
		t.Fatalf("unexpected yaml body: %q", body)
	}
}

func TestSwaggerUI_HeadAndPassthrough(t *testing.T) {
	e := echo.New()
	mw := SwaggerUIBytes([]byte("openapi: 3.0.0\n"))

	e.Use(mw)
	e.POST("/swagger", func(c echo.Context) error {
		return c.String(http.StatusAccepted, "posted")
	})

	req := httptest.NewRequest(http.MethodHead, "/swagger", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get(echo.HeaderContentType); got != contentTypeHTML {
		t.Fatalf("unexpected content type: %q", got)
	}

	if got := rec.Header().Get(echo.HeaderContentLength); got != strconv.Itoa(len(swaggerUIHTML(defaultPath))) {
		t.Fatalf("unexpected content-length: %q", got)
	}

	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body for HEAD, got %q", rec.Body.String())
	}

	reqPost := httptest.NewRequest(http.MethodPost, "/swagger", nil)
	recPost := httptest.NewRecorder()
	e.ServeHTTP(recPost, reqPost)

	if recPost.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, recPost.Code)
	}

	if recPost.Body.String() != "posted" {
		t.Fatalf("unexpected body: %q", recPost.Body.String())
	}
}

func TestSwaggerUI_ConstructorsAcceptEmptyInputs(t *testing.T) {
	tests := []struct {
		name string
		mw   echo.MiddlewareFunc
	}{
		{
			name: "bytes empty",
			mw:   SwaggerUIBytes(nil),
		},
		{
			name: "spec nil",
			mw:   SwaggerUISpec(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(tt.mw)

			reqUI := httptest.NewRequest(http.MethodGet, "/swagger", nil)
			recUI := httptest.NewRecorder()
			e.ServeHTTP(recUI, reqUI)

			if recUI.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recUI.Code)
			}

			if !strings.Contains(recUI.Body.String(), `url: "/swagger.yaml"`) {
				t.Fatalf("unexpected html body: %q", recUI.Body.String())
			}

			reqYAML := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)
			recYAML := httptest.NewRecorder()
			e.ServeHTTP(recYAML, reqYAML)

			if recYAML.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d", http.StatusOK, recYAML.Code)
			}

			if got := recYAML.Header().Get(echo.HeaderContentType); got != contentTypeYAML {
				t.Fatalf("unexpected yaml content type: %q", got)
			}

			if recYAML.Body.String() != "" {
				t.Fatalf("expected empty yaml body, got %q", recYAML.Body.String())
			}
		})
	}
}
