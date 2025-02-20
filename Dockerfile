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

# Building image
FROM alpine:latest

#Копируем основное приложение
COPY --from=builder /app/bin/sso /sso

# Копируем мигратор и миграции
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/bin/migrator /migrator

# т.к конфиг общий для мигратора и приложения копируем только один
COPY --from=builder /app/config/local.yaml ./config/local.yaml

RUN chmod +x /sso
RUN chmod +x /migrator

ENV CONFIG_PATH=config/local.yaml
ENV MIGRATIONS_PATH=migrations

EXPOSE 4000

#CMD ["./sso"]
ENTRYPOINT ["sh", "-c", "./migrator && exec ./sso"]