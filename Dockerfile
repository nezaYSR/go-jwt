# Build stage
FROM golang:1.17-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files to the container
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application to the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final stage
FROM alpine:3.14

# Install ca-certificates for SSL connections to MongoDB
# RUN apk --no-cache add ca-certificates

# Copy the built application from the previous stage
COPY --from=builder /app/app /app/app

# Install MongoDB client tools
# RUN apk add --no-cache mongodb-tools

# Set the working directory for the container
WORKDIR /app

# .env file
COPY .env ./

# Expose the port used by the application
EXPOSE 8080

# Start the application
CMD ["./app"]