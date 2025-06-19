.PHONY: build test clean run-yaml run-yaml-gzip run-yaml-custom run-yaml-input-vars run-yaml-assignment run-yaml-source test-input-param test-input-param-sub run-lowercase-example run-yaml-direct-assignment run-yaml-nested-functions

# Build the application
build:
	go build -o vibestation cmd/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f vibestation

# Run with basic YAML config
run-yaml: build
	./vibestation -config tests/configs/basic.yaml -input tests/data/sample.txt

# Run with gzip YAML config
run-yaml-gzip: build
	./vibestation -config tests/configs/gzip.yaml -input tests/data/sample.txt.gz

# Run with custom split YAML config
run-yaml-custom: build
	./vibestation -config tests/configs/custom_split.yaml -input tests/data/sample_pipe.txt

# Run with input vars YAML config
run-yaml-input-vars: build
	./vibestation -config tests/configs/input_vars_example.yaml -input tests/data/test_input_targeting.json

# Run with assignment YAML config
run-yaml-assignment: build
	./vibestation -config tests/configs/assignment_example.yaml -input tests/data/test_input_targeting.json

# Run with source key test
run-yaml-source: build
	./vibestation -config tests/configs/source_test.yaml -input tests/data/test_input_targeting.json

# Run with direct assignment example
run-yaml-direct-assignment: build
	./vibestation -config tests/configs/direct_assignment_example.yaml -input tests/data/test_input_param.json

# Run with nested functions test
run-yaml-nested-functions: build
	./vibestation -config tests/configs/nested_functions_test.yaml -input tests/data/test_input_param.json

# Install dependencies
deps:
	go mod tidy

# All-in-one: deps, test, build
all: deps test build 

# Test input parameter standardization
test-input-param: build
	@echo "Testing input parameter standardization..."
	@./vibestation -config tests/configs/input_parameter_test.yaml -input tests/data/test_input_param.json

# Test input parameter standardization with SUB sublang
test-input-param-sub: build
	@echo "Testing input parameter standardization with SUB sublang..."
	@./vibestation -config tests/configs/input_parameter_sub_test.yaml -input tests/data/test_input_param.json 

# Run lowercase example
run-lowercase-example: build
	@echo "Running lowercase example..."
	@./vibestation -config tests/configs/lowercase_example.yaml -input tests/data/test_input_param.json 
