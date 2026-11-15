FROM node:22-alpine AS web
WORKDIR /web
COPY apps/backoffice-web/package*.json ./
RUN npm ci --no-audit --no-fund
COPY apps/backoffice-web/ ./
RUN npm run build

FROM golang:1.25-alpine AS api
WORKDIR /api
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S -G app app
COPY --from=api /server /usr/local/bin/server
COPY --from=web /web/build /app/web
USER app
ENV PORT=8080 WEB_ROOT=/app/web APP_ENV=test
EXPOSE 8080
CMD ["server"]
