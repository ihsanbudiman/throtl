# ---- Build ----
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /throtl-server ./cmd/server

# ---- Run ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /throtl-server /app/throtl-server
EXPOSE 8080
ENV THROTL_PORT=8080
VOLUME ["/data"]
ENTRYPOINT ["/app/throtl-server"]
