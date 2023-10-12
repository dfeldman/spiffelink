# Use an official Go runtime as a parent image
FROM golang:1.19 AS build-env

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
ADD . /app

# Build the application
RUN go build -o spiffelink .

# Use a smaller image to run our application
# TODO THis should really be alpine-based to be consistent with SPIRE
FROM gcr.io/distroless/base-debian11

# Copy the binary from the build stage
COPY --from=build-env /app/spiffelink /spiffelink

# Run the binary when the container starts
ENTRYPOINT ["/spiffelink"]