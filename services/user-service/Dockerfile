# Use a Go base image
FROM golang:1.23.4-alpine AS build

# Set the working directory in the container
WORKDIR /app

# Copy the Go modules and source code into the container
COPY go.mod ./
RUN go mod tidy
COPY . .

# Build the Go application
RUN go build -o stockTracker .

# Use a smaller image for the final container
FROM alpine:latest  

# Set the working directory
WORKDIR /root/

# Copy the Go binary from the build image
COPY --from=build /app/stockTracker .

# Expose the port on which your microservice will run
EXPOSE 8080

# Command to run the application
CMD ["./stockTracker"]

