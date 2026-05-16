.PHONY: check-frontend check-backend check-e2e check-all

check-frontend:
	cd web && bun run type-check && bun run test:unit

check-backend:
	cd server && go test ./...

check-e2e:
	cd web && bun run test:e2e

check-all: check-frontend check-backend
