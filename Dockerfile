# Use official Go base image
FROM golang:1.22 as builder

WORKDIR /app

# Copy go mod and sum files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN go build -o jobscope .

# --------- Minimal runtime image ---------
FROM debian:bullseye-slim

WORKDIR /app

# Copy only the compiled binary and data folder
COPY --from=builder /app/jobscope .
COPY --from=builder /app/data ./data

# Make sure data folder is writable
RUN chmod -R 755 ./data

# Expose port
EXPOSE 8080

CMD ["./jobscope"]
