# Dockerfile definition for Backend application service.

# Stage 1: Build the Go application
FROM golang:1.21-alpine as build

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o main cmd/main.go ./cmd/config.go

####################################################################
# Stage 2: Create the production image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary and config file from the build stage
COPY --from=build /app/main .
COPY --from=build /app/config.yaml .

# Expose the port that the application will listen on
EXPOSE 1323

# Command to run the application
ENTRYPOINT ["./main"]
