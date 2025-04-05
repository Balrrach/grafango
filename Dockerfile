# Start with the official Golang base image
FROM golang:1.24.1-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the Go files into the container
COPY . .

# Get dependencies (if any)
RUN go mod tidy

# Build the Go app
RUN go build -o main .

# Expose port (Go app is running on port 8080)
EXPOSE 8080

# Run the Go app
CMD ["./main"]
