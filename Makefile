BINARY := smrepl
DOCKER_IMAGE_REPO := smrepl
WINDOWS=$(BINARY)_windows_amd64.exe
LINUX=$(BINARY)_linux_amd64
DARWIN=$(BINARY)_darwin_amd64
VERSION=$(shell git describe --tags --always --long --dirty)

ifdef TRAVIS_BRANCH
        BRANCH := $(TRAVIS_BRANCH)
else
        BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
endif

all: build-mac build-win build-linux
.PHONY: all

build:
	go build -o $(BINARY)
.PHONY: build

dockerbuild-go:
	docker build -t $(DOCKER_IMAGE_REPO):$(BRANCH) .
.PHONY: dockerbuild-go

build-win:
	env GOOS=windows GOARCH=amd64 go build -o $(WINDOWS)
.PHONY: build-win

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o $(LINUX)
.PHONY: build-linux

build-mac:
	env GOOS=darwin GOARCH=amd64 go build -o $(DARWIN)
.PHONY: build-mac

clean:
	rm -f $(WINDOWS) $(LINUX) $(DARWIN)

lint:
	go vet ./...
	go fmt ./...
.PHONY: lint

dockerpush: dockerbuild-go
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	docker tag $(DOCKER_IMAGE_REPO):$(BRANCH) spacemeshos/$(DOCKER_IMAGE_REPO):$(BRANCH)
	docker push spacemeshos/$(DOCKER_IMAGE_REPO):$(BRANCH)
.PHONY: dockerpush
