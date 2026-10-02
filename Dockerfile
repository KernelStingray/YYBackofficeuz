FROM golang:alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Argument for the service name
ARG SERVICE_NAME
ENV SERVICE_NAME=${SERVICE_NAME}

# Build the Go app with limited parallelism to save memory during concurrent builds
RUN CGO_ENABLED=0 GOOS=linux go build -p 2 -installsuffix cgo -o main ./cmd/${SERVICE_NAME}

######## Start a new stage from scratch #######
FROM alpine:latest  

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Default env vars
ENV TZ=UTC
ENV GIN_MODE=release


# Keys and .env config are expected to be mounted via docker-compose volumes

# Command to run the executable
CMD ["./main"]
