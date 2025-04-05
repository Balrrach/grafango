FROM golang:1.24.1-alpine

# Set the working directory that matches the module name
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the application
RUN go build -o /app/bin/grafango ./cmd/grafango

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/bin/grafango"]