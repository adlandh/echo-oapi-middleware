# AGENTS.md

## Repo Shape
- Single Go module: `github.com/adlandh/echo-oapi-middleware/v2`, package `echooapimiddleware` at repo root; there are no subpackages, commands, or generated sources.
- Public entrypoints are `SwaggerYaml`, `SwaggerYamlWithConfig`, `SwaggerUI`, and `SwaggerUIWithConfig` in `swaggeryaml.go` and `swaggerui.go`.
- This repo uses Echo v5 (`github.com/labstack/echo/v5`); handlers/tests use `func(c *echo.Context) error`, not the Echo v4 `echo.Context` value signature.

## Commands
- Full test command from CI: `go test -race -coverprofile=coverage.txt -covermode=atomic ./...`.
- Fast local test loop: `go test ./...`.
- Focus one test: `go test -run TestSwaggerUI_DefaultPaths .` or replace the test name.
- Run the benchmark: `go test -bench BenchmarkSwaggerYamlGET -benchmem .`.
- CI lint overwrites the checked-in `.golangci.yml` before running: `curl -sS https://raw.githubusercontent.com/adlandh/golangci-lint-config/refs/heads/main/.golangci.yml -o .golangci.yml`, then `golangci-lint run`. Do not assume edits to the local `.golangci.yml` affect CI unless that workflow changes.

## Middleware Behavior To Preserve
- Specs are serialized once when middleware is created, not per request.
- `SwaggerYaml` serves only `GET` and `HEAD` at `/swagger.yaml` by default; other methods and paths must pass through to the next handler.
- `SwaggerUI` serves UI at `/swagger`, `/swagger/`, and `/swagger/index.html` by default, and also wires the YAML endpoint.
- `HEAD` responses must set `Content-Length` and return no body for both YAML and UI responses.
- `nil` specs are accepted and produce successful empty YAML responses.
- `KeepServers` defaults to false: YAML output strips `servers`, but must not mutate the caller's `*openapi3.T`.
- Swagger UI assets are referenced from the unpkg `swagger-ui-dist@5` CDN in generated HTML.
