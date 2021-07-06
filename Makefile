.PHONY: all test clean

RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m

all: clean build test

build:
	@echo "${RED}Building decode.go${NC}"
	go build -o ./bin/decode ./cmd/decode/decode.go
	@echo "${RED}Building encode.go${NC}"
	go build -o ./bin/encode ./cmd/encode/encode.go
	@echo "${GREEN}See binaries in ./bin${NC}"

test:
	@echo "${YELLOW}Testing${NC}"
	go test ./test -v

clean:
	@echo "${RED}Deleting old binaries${NC}"
	rm -rf ./bin
