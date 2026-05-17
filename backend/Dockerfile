# ---- Build ----
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /shareai-server ./cmd/server

# ---- Run ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates libc6-compat
WORKDIR /app
COPY --from=builder /shareai-server /app/shareai-server
EXPOSE 8080
ENV SHAREAI_PORT=8080
VOLUME ["/data"]
ENTRYPOINT ["/app/shareai-server"]
