FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o vikunja_mcp main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/vikunja_mcp /usr/local/bin/vikunja_mcp

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["vikunja_mcp"]
