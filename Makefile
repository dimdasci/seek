# Basic variables
BINARY_NAME=seek
BINARY_DIR=bin

# Version information
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH=$(shell git rev-parse --short HEAD)

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.CommitHash=${COMMIT_HASH}"

.PHONY: build clean

# Build the application
build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BINARY_DIR}
	go build ${LDFLAGS} -o ${BINARY_DIR}/${BINARY_NAME} .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf ${BINARY_DIR}

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_DIR}/${BINARY_NAME}-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_DIR}/${BINARY_NAME}-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_DIR}/${BINARY_NAME}-windows-amd64.exe .
