FROM golang:1.20-alpine
WORKDIR /app
COPY main.go .
RUN go mod init vault_init && go mod tidy && go build -o vault_init main.go
CMD ["./vault_init"]
