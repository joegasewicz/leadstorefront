FROM node:22-alpine AS assets

WORKDIR /src/platform_web

COPY platform_web/package*.json ./
RUN npm ci

COPY platform_web/src ./src
COPY platform_web/tsconfig.json platform_web/vite.config.ts ./
RUN npm run build

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkgs ./pkgs
COPY platform_web/routes ./platform_web/routes

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/platform_web ./cmd/platform_web

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/platform_web ./platform_web_server
COPY --from=assets /src/platform_web/static/assets ./platform_web/static/assets
COPY platform_web/templates ./platform_web/templates

RUN mkdir -p uploads

ENV WEB_ADDR=:8000

EXPOSE 8000

CMD ["./platform_web_server"]
