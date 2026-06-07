FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN rm -f go.mod go.sum
RUN go mod init myapp && \
    go mod edit -go=1.22 && \
    go get golang.org/x/crypto@v0.17.0 && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o invoicer .
FROM scratch
COPY --from=builder /app/invoicer /invoicer
COPY --from=builder /app/cert.pem /cert.pem
COPY --from=builder /app/key.pem /key.pem
USER 65534:65534
EXPOSE 8080 8443
ENTRYPOINT ["/invoicer"]
