# Use a Go base image
FROM golang:1.23.4-alpine AS build

# Set the working directory in the container
WORKDIR /workspace
COPY . .

# Copy the Go modules and source code into the container
WORKDIR /workspace/services/user-service
RUN go build -o /usr/local/bin/user-service .


# Build the Go application
RUN go build -o stockTracker .

# Use a smaller image for the final container
FROM alpine:latest  

COPY --from=build /usr/local/bin/user-service /root/user-service

# Set the working directory
WORKDIR /root

# Expose the port on which your microservice will run
EXPOSE 8080

# Command to run the application
CMD ["./user-service"]

