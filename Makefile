# Local mirror of CI checks — run `make check` before pushing (see CLAUDE.md "Git Workflow").
# Blocking CI gates: go vet, go test, Docker smoke builds, Trivy, CodeQL.
# Lint is advisory in CI (continue-on-error) while imported-codebase lint debt is paid down.

GOLANGCI_LINT_IMAGE := golangci/golangci-lint:latest

.PHONY: check vet test lint frontend smoke

check: vet test frontend lint

vet:
	cd backend && go vet ./...

test:
	cd backend && go test ./...

frontend:
	cd frontend && { [ -d node_modules ] || npm ci; } && npm run build

# Advisory: reports issues but does not fail `make check`, matching CI
lint:
	-docker run --rm -v $(CURDIR)/backend:/app -w /app $(GOLANGCI_LINT_IMAGE) golangci-lint run ./...
	@echo "lint is advisory (non-blocking in CI) while lint debt is paid down"

# Full CI parity for the PR smoke test: Docker image builds (slow)
smoke:
	docker build -t pr-smoke-backend ./backend
	docker build -t pr-smoke-frontend ./frontend
