.PHONY: build test clean run run-gzip run-yaml run-yaml-gzip run-yaml-custom run-yaml-input-vars run-yaml-assignment run-yaml-source test-input-param test-input-param-sub run-lowercase-example

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

# Run with basic YAML config
run-yaml: build
	./vibestation -config configs/basic.yaml -input sample.txt

# Run with gzip YAML config
run-yaml-gzip: build
	./vibestation -config configs/gzip.yaml -input sample.txt.gz

# Run with custom split YAML config
run-yaml-custom: build
	./vibestation -config configs/custom_split.yaml -input sample_pipe.txt

# Run with input vars YAML config
run-yaml-input-vars: build
	./vibestation -config configs/input_vars_example.yaml -input test_input_targeting.json

# Run with assignment YAML config
run-yaml-assignment: build
	./vibestation -config configs/assignment_example.yaml -input test_input_targeting.json

# Run with source key test
run-yaml-source: build
	./vibestation -config configs/source_test.yaml -input test_input_targeting.json

# Install dependencies
deps:
	go mod tidy

# All-in-one: deps, test, build
all: deps test build 

# Test input parameter standardization
test-input-param: build
	@echo "Testing input parameter standardization..."
	@./vibestation -config configs/input_parameter_test.yaml -input test_input_param.json

# Test input parameter standardization with SUB DSL
test-input-param-sub: build
	@echo "Testing input parameter standardization with SUB DSL..."
	@./vibestation -config configs/input_parameter_sub_test.yaml -input test_input_param.json 

# Run lowercase example
run-lowercase-example: build
	@echo "Running lowercase example..."
	@./vibestation -config configs/lowercase_example.yaml -input test_input_param.json 
