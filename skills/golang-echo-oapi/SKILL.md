---
name: golang-echo-oapi
description: Use when working on github.com/adlandh/echo-oapi-middleware/v2, an Echo v5 middleware library that serves OpenAPI YAML and Swagger UI from kin-openapi specs.
---

# golang-echo-oapi

## When To Use
- Use this skill for changes to `github.com/adlandh/echo-oapi-middleware/v2` or code that consumes this library.
- Apply it when touching `SwaggerYaml`, `SwaggerYamlWithConfig`, `SwaggerUI`, `SwaggerUIWithConfig`, their tests, or README examples.

## Library Shape
- Package name is `echooapimiddleware`; it is a single-package Go module with no generated code or subpackages.
- It targets Go 1.25 and Echo v5 (`github.com/labstack/echo/v5`). Echo handlers use `func(c *echo.Context) error`.
- Public API accepts `*openapi3.T` from `github.com/getkin/kin-openapi/openapi3`; do not add runtime parsing as part of the middleware path unless explicitly requested.

## Behavior To Preserve
- Serialize the OpenAPI spec once when middleware is created, not on every request.
- `SwaggerYaml` defaults to `/swagger.yaml` and serves only `GET` and `HEAD`; all other methods and paths pass through to the next handler.
- `SwaggerUI` defaults to `/swagger`, also serves `/swagger/` and `/swagger/index.html`, and wires the YAML endpoint.
- `HEAD` responses for YAML and UI must set `Content-Length` and write no body.
- `nil` specs are valid and produce successful empty YAML responses.
- `KeepServers` defaults to false: omit top-level `servers` from emitted YAML without mutating the caller's `*openapi3.T`.
- Generated Swagger UI HTML references unpkg `swagger-ui-dist@5` assets.

## Verification Commands
- Fast local check: `go test ./...`.
- CI-equivalent tests: `go test -race -coverprofile=coverage.txt -covermode=atomic ./...`.
- Focused tests: `go test -run TestSwaggerYaml_RequestRouting .` or `go test -run TestSwaggerUI_DefaultPaths .`.
- Benchmark YAML serving path: `go test -bench BenchmarkSwaggerYamlGET -benchmem .`.
- CI lint downloads the shared config before running: `curl -sS https://raw.githubusercontent.com/adlandh/golangci-lint-config/refs/heads/main/.golangci.yml -o .golangci.yml`, then `golangci-lint run`.

## Implementation Notes
- Keep the middleware chain passthrough behavior explicit; tests depend on non-matching requests reaching downstream Echo routes.
- Prefer small direct functions over new abstractions; this library is intentionally tiny.
- If changing config defaults or route matching, update both README examples and tests in the same change.

## Integrating Into A Project
- **Do not try to embed or parse `swagger.yaml` / `openapi.yaml` at runtime.** This library accepts `*openapi3.T` from `github.com/getkin/kin-openapi/openapi3`, not a raw YAML file.
- Before reaching for `//go:embed` or `os.ReadFile`, check the generated code — most OpenAPI code generators (e.g. `oapi-codegen`) produce a `GetSpec` function that returns the parsed `*openapi3.T` directly. Pass that to `SwaggerYAML` or `SwaggerUI`.
