# Build Step
FROM golang:1.15.4-alpine3.12 AS builder

# Dependencies
RUN apk update && apk add --no-cache upx

# Source
WORKDIR $GOPATH/src/github.com/Depado/capybara
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify
COPY . .

# Build
ARG build
ARG version
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=${version} -X main.Build=${build}" -o /tmp/capybara
RUN upx /tmp/capybara


# Final Step
FROM gcr.io/distroless/static
COPY --from=builder /tmp/capybara /go/bin/capybara

VOLUME [ "/data" ]
WORKDIR /data
ENTRYPOINT ["/go/bin/capybara"]
