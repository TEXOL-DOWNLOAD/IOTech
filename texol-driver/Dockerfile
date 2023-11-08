# syntax=docker/dockerfile:1

FROM golang:1.21-alpine3.18

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod ./
COPY *.go ./

RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /texol-ble-driver

# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can (optionally) document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 8888

# Run
CMD [ "/texol-ble-driver" ]
