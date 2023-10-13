# Use an official Go runtime as a parent image
FROM golang:1.19 AS build-env

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
ADD . /app

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o spiffelink .

# We put the spiffe-link in the spire-agent container so they can share a socket easily
FROM ghcr.io/spiffe/spire-agent:1.8.1

# Copy the binary from the build stage
COPY --from=build-env /app/spiffelink /bin/spiffelink

# Run the binary when the container starts
ENTRYPOINT ["/bin/spiffelink"]