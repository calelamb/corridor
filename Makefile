.PHONY: generate check-generated test web-check

generate:
	go tool sqlc generate
	go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

check-generated: generate
	git ls-files --error-unmatch internal/api/generated.go internal/db/generated >/dev/null
	git diff --exit-code -- internal/api/generated.go internal/db/generated
	git diff --cached --exit-code -- internal/api/generated.go internal/db/generated

test:
	go test -race ./...

web-check:
	npm --prefix web run check
	npm --prefix web run lint
