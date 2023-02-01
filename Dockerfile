# Build Step
FROM golang:1.17.3-alpine3.14 AS builder

# Dependencies
RUN apk update && apk add --no-cache upx make git

# Source
WORKDIR $GOPATH/src/github.com/Depado/capybara
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
RUN make tmp
RUN upx --best --lzma /tmp/capybara

# Final Step
FROM gcr.io/distroless/static
COPY --from=builder /tmp/capybara /go/bin/capybara
VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/capybara"]