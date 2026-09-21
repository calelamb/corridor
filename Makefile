.PHONY: generate check-generated test web-check

generate:
	go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

check-generated: generate
	git ls-files --error-unmatch internal/api/generated.go >/dev/null
	git diff --exit-code -- internal/api/generated.go
	git diff --cached --exit-code -- internal/api/generated.go

test:
	go test -race ./...

web-check:
	npm --prefix web run check
	npm --prefix web run lint
