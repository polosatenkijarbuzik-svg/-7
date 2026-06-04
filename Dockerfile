# Этап сборки
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o invoicer .

# Финальный образ — минимальный scratch
FROM scratch
COPY --from=builder /app/invoicer /invoicer
USER 65534:65534   # nobody
EXPOSE 8080
ENTRYPOINT ["/invoicer"]
