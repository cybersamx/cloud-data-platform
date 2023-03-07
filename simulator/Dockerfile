### Build Stage ###

FROM golang:1.19.2-alpine as build

# Install system dependencies for the build
RUN apk add --no-cache \
    ca-certificates \
    git \
    make

# Set the build environments
WORKDIR /go/src

# Install build dependencies
COPY Makefile ./
COPY go.mod go.sum ./
RUN make install

# Copy the project and build
COPY . ./
RUN make build

### Final Stage ###

FROM alpine

# Install tini - need for production for graceful shutdowns
# TODO: Check if tini is still needed with newer os.
RUN apk add --no-cache tini

# Install system runtime dependencies
RUN apk add --no-cache \
    ca-certificates

# Copy binaries and config file over
WORKDIR /go/bin
USER nobody:nobody
COPY --from=build /go/src/bin/simulator ./

# Execute
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/go/bin/simulator", "start"]
