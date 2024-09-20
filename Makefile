package := $(shell basename `pwd`)

# strip debugging information & include buildnumber in executable
LDFLAGS=-ldflags "-s -w -X main.buildnum=${BUILD_NUM}"

.PHONY: default get codetest build release fmt lint vet run

default: fmt codetest

get:
ifneq ("$(CI)", "true")
	go get -u ./...
	go mod tidy
endif
	go mod download
	go mod verify

codetest: lint vet

build:
	mkdir -p target
	rm -f target/*
	GOOS=linux GOARCH=amd64 go build -tags=nomsgpack ${LDFLAGS} -v -o target/$(package) github.com/MikeAlbertFleetSolutions/paycor-driver-sync/cmd/paycor-driver-sync

release: build
	zip -j target/$(package)_linux_amd64.zip target/$(package)
	rm -f target/$(package)

fmt:
	go fmt ./...

lint:
ifeq ("$(CI)", "true")
	$(shell go env GOPATH)/bin/golangci-lint run --verbose --timeout 3m
else
	$(shell go env GOPATH)/bin/golangci-lint run --fix
endif

vet:
	go vet -all ./...

run: default build
	target/paycor-driver-sync -config paycor-driver-sync.yaml
