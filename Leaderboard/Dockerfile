# Build Stage
FROM golang:1.20-alpine AS builder
RUN apk update && apk add --no-cache gcc musl-dev

ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o leaderboard .

# Run stage

FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates

ENV PORT=8080

WORKDIR /root/

COPY --from=builder /app/leaderboard .
COPY --from=builder /app/static ./static
COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./leaderboard"]