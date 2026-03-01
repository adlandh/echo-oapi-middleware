package echooapimiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

func TestSwaggerUI_DefaultPaths(t *testing.T) {
	e := echo.New()
	mw := SwaggerUI(&openapi3.T{
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

func TestSwaggerUI_ConstructorsAcceptEmptyInputs(t *testing.T) {
	e := echo.New()
	e.Use(SwaggerUI(nil))

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
}
