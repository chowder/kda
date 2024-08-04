FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o kda -ldflags="-s -w"

FROM alpine

COPY --from=builder /build/kda /app/kda

ENTRYPOINT ["/app/kda"]