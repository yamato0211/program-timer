FROM golang:1.20-buster AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o bot cmd/bot/main.go

FROM gcr.io/distroless/base-debian11:latest

WORKDIR /

COPY --from=builder /app/bot bot
COPY --from=builder /app/.env .env

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["./bot"]

