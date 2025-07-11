# FROM golang:1.21
FROM golang:1.24

# Set the working directory.
WORKDIR /app

# Download the Go modules.
COPY  go.mod go.sum ./
RUN go mod download

# Copy the source code files.
COPY *.go ./

# Compile the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-portal-microcks-integration

# Expose a port if needed.
# This application does not need to expose anything, so we can simply comment this out.
#EXPOSE 8080

# Run
CMD ["/go-portal-microcks-integration"]