# Build stage
FROM golang:1.26.0-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o finalcountdown .

# Runtime stage — minimal scratch image
FROM scratch

COPY --from=builder /build/finalcountdown /finalcountdown
COPY static/ /static/

EXPOSE 8080

ENTRYPOINT ["/finalcountdown"]
