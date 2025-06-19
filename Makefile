.PHONY: build test clean run run-gzip run-config run-config-gzip run-config-custom

# Build the application
build:
	go build -o vibestation cmd/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f vibestation

# Run with sample text file (legacy mode)
run: build
	./vibestation -input sample.txt

# Run with gzipped sample file (legacy mode)
run-gzip: build
	./vibestation -input sample.txt.gz -gzip

# Run with basic JSON config
run-config: build
	./vibestation -config configs/basic.json -input sample.txt

# Run with gzip JSON config
run-config-gzip: build
	./vibestation -config configs/gzip.json -input sample.txt.gz

# Run with custom split JSON config
run-config-custom: build
	./vibestation -config configs/custom_split.json -input sample_pipe.txt

# Install dependencies
deps:
	go mod tidy

# All-in-one: deps, test, build
all: deps test build 
