.DEFAULT_GOAL := build

CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags 2> /dev/null || echo "0.1.0")
BUILD=$(shell git rev-parse HEAD 2> /dev/null || echo "undefined")
BUILDDATE=$(shell LANG=en_us_88591 date)
BINARY=capybara
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD) -s -w"
LDFLAGS=-ldflags "-X 'github.com/depado/capybara/cmd.Version=$(VERSION)' \
		-X 'github.com/depado/capybara/cmd.Build=$(BUILD)' \
		-X 'github.com/depado/capybara/cmd.Time=$(BUILDDATE)' -s -w"
PACKEDFLAGS=-ldflags "-X 'github.com/depado/capybara/cmd.Version=$(VERSION)' \
		-X 'github.com/depado/capybara/cmd.Build=$(BUILD)' \
		-X 'github.com/depado/capybara/cmd.Time=$(BUILDDATE)' \
		-X 'github.com/depado/capybara/cmd.Packer=upx --best --lzma' -s -w"

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build-noproto
build-noproto: ## Build without regenerating the go grpc bindings
	go build $(LDFLAGS) -o $(BINARY)

.PHONY: build
build: proto ## Build
	go build $(LDFLAGS) -o $(BINARY)

.PHONY: packed
packed: proto ## Build a packed version
	go build $(PACKEDFLAGS) -o $(BINARY)
	upx --best --lzma $(BINARY)

.PHONY: proto
proto: ## Generate protobuf
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pb/capybara.proto
	protoc --go_out=. --go_opt=paths=source_relative ./pb/database.proto

.PHONY: docker
docker: proto ## Build the docker image
	docker build -t $(BINARY):latest -t $(BINARY):$(BUILD) -f Dockerfile .

.PHONY: tmp
tmp: ## Build and output the binary in /tmp
	go build $(LDFLAGS) -o /tmp/$(BINARY)

.PHONY: release
release: ## Create a new release on Github
	goreleaser

.PHONY: snapshot
snapshot: ## Create a new snapshot release
	goreleaser --snapshot --clean

.PHONY: lint
lint: ## Runs the linter
	$(GOPATH)/bin/golangci-lint run --exclude-use-default=false

.PHONY: test
test: ## Run the test suite
	CGO_ENABLED=1 go test -race -coverprofile="coverage.txt" ./...

.PHONY: clean
clean: ## Remove the binary
	if [ -f $(BINARY) ] ; then rm $(BINARY) ; fi
	if [ -f coverage.txt ] ; then rm coverage.txt ; fi
