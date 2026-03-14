package echooapimiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v5"
)

func TestSwaggerUI_DefaultPaths(t *testing.T) {
	e := echo.New()
	mw := SwaggerUI(&openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "API", Version: "1.0.0"},
	})

	e.Use(mw)
	e.GET("/users", func(c *echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", http.NoBody)
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

func TestSwaggerUI_ConstructorsAcceptEmptyInputs(t *testing.T) {
	e := echo.New()
	e.Use(SwaggerUI(nil))

	reqUI := httptest.NewRequest(http.MethodGet, "/swagger", http.NoBody)
	recUI := httptest.NewRecorder()
	e.ServeHTTP(recUI, reqUI)

	if recUI.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recUI.Code)
	}

	if !strings.Contains(recUI.Body.String(), `url: "/swagger.yaml"`) {
		t.Fatalf("unexpected html body: %q", recUI.Body.String())
	}

	reqYAML := httptest.NewRequest(http.MethodGet, "/swagger.yaml", http.NoBody)
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
}

func TestSwaggerUI_CustomPathHeadAndPassthrough(t *testing.T) {
	e := echo.New()
	e.Use(SwaggerUIWithConfig(&openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "API", Version: "1.0.0"},
	}, SwaggerUIConfig{
		Path:     "/docs",
		SpecPath: "/openapi.yaml",
	}))
	e.GET("/users", func(c *echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	reqHead := httptest.NewRequest(http.MethodHead, "/docs/index.html", http.NoBody)
	recHead := httptest.NewRecorder()
	e.ServeHTTP(recHead, reqHead)

	if recHead.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recHead.Code)
	}

	if got := recHead.Header().Get(echo.HeaderContentType); got != contentTypeHTML {
		t.Fatalf("unexpected html content type: %q", got)
	}

	if got := recHead.Header().Get(echo.HeaderContentLength); got == "" || got == "0" {
		t.Fatalf("unexpected content-length: %q", got)
	}

	if recHead.Body.Len() != 0 {
		t.Fatalf("expected empty HEAD body, got %q", recHead.Body.String())
	}

	reqPassthrough := httptest.NewRequest(http.MethodPost, "/docs", http.NoBody)
	recPassthrough := httptest.NewRecorder()
	e.ServeHTTP(recPassthrough, reqPassthrough)

	if recPassthrough.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recPassthrough.Code)
	}

	reqRoute := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	recRoute := httptest.NewRecorder()
	e.ServeHTTP(recRoute, reqRoute)

	if recRoute.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recRoute.Code)
	}

	if recRoute.Body.String() != "users" {
		t.Fatalf("unexpected passthrough body: %q", recRoute.Body.String())
	}
}
