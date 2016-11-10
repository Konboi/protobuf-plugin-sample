TEST_FILE = $(shell glide novendor)

test: deps

	go test -v ${TEST_FILE}

build: deps

	go build -o app

deps:

	go get github.com/golang/lint/golint
	glide install

debug: build

	protoc --plugin=protoc-gen-custom=./app player.proto ping.proto --custom_out=pb/
	gofmt -w -d tmp/
