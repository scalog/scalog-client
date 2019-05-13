# Start from golang v1.11 base image
FROM golang:1.11 as builder

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . /go/src/github.com/scalog/scalog-client/

# Set the Current Working Directory inside the container
WORKDIR /go/src/github.com/scalog/scalog-client/

# Download dependencies
RUN set -x && \
    go get github.com/golang/dep/cmd/dep && \
    dep ensure -v

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags \
    '-extldflags "-static"' -o scalog-client .

######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/src/github.com/scalog/scalog-client/scalog-client /app/

WORKDIR /app
