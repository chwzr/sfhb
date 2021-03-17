############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/sfhb/
COPY . .
# Fetch dependencies.
# Using go get.
RUN cd src && go get -d -v
# Build the binary.
RUN go build -o /go/bin/sfhb src/main.go

############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/sfhb /go/bin/sfhb
# Run the hello binary.
ENTRYPOINT ["/go/bin/sfhb"]