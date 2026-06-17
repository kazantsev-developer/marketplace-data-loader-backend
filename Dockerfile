FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /build/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /build/sync ./cmd/sync

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /build/server /build/sync ./
COPY migrations/ ./migrations/

EXPOSE 3000

CMD ["./server"]