# Build stage
FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/ssh-battos .

# Runtime stage
FROM alpine:3.19
RUN adduser -D -u 10001 battos
WORKDIR /app
COPY --from=build /out/ssh-battos .
RUN mkdir -p /app/.ssh && chown -R battos:battos /app
USER battos
ENV SSH_HOST=0.0.0.0 \
    SSH_PORT=23234 \
    SSH_HOST_KEY=/app/.ssh/id_ed25519
EXPOSE 23234
ENTRYPOINT ["/app/ssh-battos"]
