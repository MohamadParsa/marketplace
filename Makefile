.PHONY: api-test api-run web-install web-build

api-test:
	go test ./...

api-run:
	go run ./cmd/api

web-install:
	npm install --legacy-peer-deps --no-audit --no-fund --prefix apps/backoffice-web

web-build:
	npm run build --prefix apps/backoffice-web
