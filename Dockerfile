FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o tsuna .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/tsuna /usr/local/bin/tsuna

EXPOSE ${PORT:-8080}
CMD ["tsuna", "signal"]
