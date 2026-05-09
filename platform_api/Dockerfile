FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkgs ./pkgs
COPY platform_api ./platform_api

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/platform_api ./cmd/platform_api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/platform_api ./platform_api

RUN mkdir -p uploads

ENV API_ADDR=:8001

EXPOSE 8001

CMD ["./platform_api"]
