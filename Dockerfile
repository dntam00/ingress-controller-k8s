FROM golang:1.23-bookworm as builder

# Create and change to the app directory.
WORKDIR /app

# RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
#     ca-certificates && update-ca-certificates && \
#     rm -rf /var/lib/apt/lists/*

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -v -o bin ./cmd/main.go

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && update-ca-certificates 2>/dev/null || true && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/bin /app/bin
#RUN MKDIR -p /root/.kube
COPY --from=builder /app/config /root/.kube/config

# Run the web service on container startup.
CMD ["/app/bin"]