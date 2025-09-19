# Dockerfile

# --- Stage 1: Builder ---
# Use the official Go image with an Alpine Linux base. Alpine is a lightweight
# Linux distribution, which helps keep the build stage smaller.
FROM golang:1.22-alpine AS builder

# Add build argument for commit SHA for traceability
ARG COMMIT_SHA=unknown

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies.
# This is done as a separate step to leverage Docker's layer caching.
# Dependencies will only be re-downloaded if these files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application.
# - CGO_ENABLED=0 disables CGO, creating a statically linked binary.
# -o main specifies the output file name.
# -ldflags="-s -w" strips debugging information, reducing the binary size.
# -X main.CommitSHA embeds the commit SHA into the binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.CommitSHA=${COMMIT_SHA}" -o main .

# --- Stage 2: Final Image ---
# Use a minimal Alpine image for the final stage. It's much smaller than
# the Go development image.
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Alpine Linux does not include root CA certificates by default.
# These are necessary for making secure HTTPS connections to other services.
# It's a best practice to include them.
RUN apk --no-cache add ca-certificates

# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/main .

# The application will read the CSV file from a /data directory.
# This directory will be provided by a Docker volume at runtime.
# We don't create it here, but it's good practice to document its expected presence.

# Expose the port the application will run on.
EXPOSE 8080

# The command to run when the container starts.
CMD ["./main"]