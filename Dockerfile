# Build Step
FROM golang:1.24.0-alpine@sha256:beded3aa2b820de62cd378834292360140504b2a12c9544d4f2b7523237a8b8d as builder

# Dependencies
RUN apk update && apk add --no-cache make git

# Source
WORKDIR $GOPATH/src/github.com/depado/capybara
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
RUN make tmp

# Final Step
FROM gcr.io/distroless/static@sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac
COPY --from=builder /tmp/capybara /go/bin/capybara
VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/capybara"]
