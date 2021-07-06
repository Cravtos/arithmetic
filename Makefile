RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m

ORIGINAL=./tests/data/compress_me.txt
ENCODED=./tests/data/compress_me.encoded
DECODED=./tests/data/compress_me.decoded

all: clean build test

build:
	mkdir -p ./bin
	@echo "${RED}Building decode.go${NC}"
	go build -o ./bin/decode ./cmd/decode/decode.go
	@echo "${RED}Building encode.go${NC}"
	go build -o ./bin/encode ./cmd/encode/encode.go
	@echo "${GREEN}See binaries in ./bin${NC}"

test:
	@echo "${YELLOW}Testing on compress_me${NC}"
	@echo "${RED}Encoding${NC}"
	./bin/encode -input ${ORIGINAL} -output ${ENCODED}
	@echo "${RED}Decoding${NC}"
	./bin/decode -input ${ENCODED} -output ${DECODED}

	@echo "${YELLOW}Looking at differences...${NC}"
	@if cmp -s "${ORIGINAL}" "${DECODED}"; then \
		echo "${GREEN}Test passed!${NC}"; \
	else \
		echo "${RED}Test failed!${NC}"; \
	fi

clean:
	@echo "${RED}Deleting old binaries${NC}"
	rm -rf ./bin
	@echo "${RED}Deleting resulting test files${NC}"
	rm -rf ${ENCODED}
	rm -rf ${DECODED}
