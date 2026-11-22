FROM node:22-alpine AS web
WORKDIR /repo
COPY apps/backoffice-web/package*.json apps/backoffice-web/
COPY apps/storefront-web/package*.json apps/storefront-web/
RUN npm --prefix apps/backoffice-web ci --no-audit --no-fund && npm --prefix apps/storefront-web ci --no-audit --no-fund
COPY apps/backoffice-web/ apps/backoffice-web/
COPY apps/storefront-web/ apps/storefront-web/
COPY packages/commerce-client/ packages/commerce-client/
RUN npm --prefix apps/backoffice-web run build && npm --prefix apps/storefront-web run build

FROM golang:1.25-alpine AS api
WORKDIR /api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/api && CGO_ENABLED=0 go build -trimpath -o /migrate ./cmd/migrate && CGO_ENABLED=0 go build -trimpath -o /bootstrap-admin ./cmd/bootstrap-admin && CGO_ENABLED=0 go build -trimpath -o /catalog-import ./cmd/catalog-import

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S -G app app
COPY --from=api /server /usr/local/bin/server
COPY --from=api /migrate /usr/local/bin/migrate
COPY --from=api /bootstrap-admin /usr/local/bin/bootstrap-admin
COPY --from=api /catalog-import /usr/local/bin/catalog-import
COPY --from=web /repo/apps/backoffice-web/build /app/backoffice
COPY --from=web /repo/apps/storefront-web/build /app/storefront
COPY db/migrations/ /app/migrations/
USER app
ENV PORT=8080 WEB_ROOT=/app/backoffice WEB_SURFACE=backoffice APP_ENV=production
EXPOSE 8080
CMD ["server"]
