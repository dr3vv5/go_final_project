FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/main .

COPY --from=builder /app/web ./web

EXPOSE 7540

CMD ["./main"]