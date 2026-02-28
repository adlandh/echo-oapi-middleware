package echooapimiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

func BenchmarkSwaggerYamlGET(b *testing.B) {
	b.Run("bytes", func(b *testing.B) {
		e := echo.New()
		mw := SwaggerYamlBytes([]byte("openapi: 3.0.0\ninfo:\n  title: API\n  version: 1.0.0\n"))

		e.Use(mw)
		e.GET("/users", func(c echo.Context) error {
			return c.String(http.StatusOK, "users")
		})

		req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				b.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
		}
	})

	b.Run("spec", func(b *testing.B) {
		e := echo.New()
		mw := SwaggerYamlSpec(&openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:   "API",
				Version: "1.0.0",
			},
			Paths: &openapi3.Paths{},
		})

		e.Use(mw)
		e.GET("/users", func(c echo.Context) error {
			return c.String(http.StatusOK, "users")
		})

		req := httptest.NewRequest(http.MethodGet, "/swagger.yaml", nil)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				b.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
		}
	})
}
