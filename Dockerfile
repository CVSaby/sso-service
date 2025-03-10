FROM golang:latest AS builder

RUN mkdir /app
WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go build -ldflags="-extldflags=-static" -o ./bin/sso ./cmd/sso/main.go

RUN go build -ldflags="-extldflags=-static" -o ./bin/migrator ./cmd/migrator/main.go

FROM alpine:latest

COPY --from=builder /app/bin/sso /sso

# мигратор и миграции
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/bin/migrator /migrator

# Конфиг
COPY --from=builder /app/config/local.yaml ./config/local.yaml

RUN chmod +x /sso
RUN chmod +x /migrator

ENV CONFIG_TYPE="file"
ENV CONFIG_PATH=config/local.yaml

ENV MIGRATIONS_PATH=migrations

EXPOSE 4000
ENTRYPOINT ["sh", "-c", "./migrator && exec ./sso"]