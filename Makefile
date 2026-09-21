.PHONY: generate check-generated test web-check web-build verify integration security images coverage format-check

generate:
	go tool sqlc generate
	go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

check-generated: generate
	git ls-files --error-unmatch internal/api/generated.go internal/db/generated/db.go internal/db/generated/models.go internal/db/generated/coverage.sql.go >/dev/null
	git diff --exit-code -- internal/api/generated.go internal/db/generated
	git diff --cached --exit-code -- internal/api/generated.go internal/db/generated
	@test -z "$$(git ls-files --others --exclude-standard -- internal/db/generated)"

web-build:
	npm --prefix web run build

web-check:
	npm --prefix web run check
	npm --prefix web run lint
	npm --prefix web run test:unit -- --run --coverage

test: web-build
	go test -race ./...

format-check:
	@test -z "$$(gofmt -l cmd internal tests)"
	@test -z "$$(go tool goimports -l cmd internal tests)"
	go vet ./...

images:
	docker build -f deploy/postgres.Dockerfile --target postgis -t corridor-postgres-no-h3:foundation .
	docker build -f deploy/postgres.Dockerfile -t corridor-postgres:foundation .
	docker build -f deploy/minio.Dockerfile -t corridor-minio:foundation .

integration:
	go test -race -tags=integration ./internal/db ./internal/storage ./internal/cli -count=1

coverage:
	mkdir -p coverage
	go test -race -tags=integration -coverpkg=./internal/... -coverprofile=coverage/go.out ./internal/...
	awk -f scripts/coverage.awk coverage/go.out

security:
	go tool gosec -exclude-generated ./...
	go tool govulncheck ./...
	npm --prefix web audit --audit-level=low

verify: check-generated web-check web-build format-check test coverage security
