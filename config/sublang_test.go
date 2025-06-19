package config

import (
	"testing"
)

func TestParserBasic(t *testing.T) {
	parser := NewParser()
	sub := `split_string(separator="\n")
send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Check first config (split)
	if configs[0]["type"] != "split_string" {
		t.Errorf("Expected type 'split_string', got '%s'", configs[0]["type"])
	}
	if configs[0]["separator"] != "\n" {
		sep, _ := configs[0]["separator"].(string)
		t.Errorf("Expected separator (newline, byte 10), got '%v' (bytes: %v)", sep, []byte(sep))
	}

	// Check second config (print)
	if configs[1]["type"] != "send_stdout" {
		t.Errorf("Expected type 'send_stdout', got '%s'", configs[1]["type"])
	}
}

func TestParserGzip(t *testing.T) {
	parser := NewParser()
	sub := `decompress_gzip()
split_string(separator="\n")
send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check first config (gzip_decompress)
	if configs[0]["type"] != "decompress_gzip" {
		t.Errorf("Expected type 'decompress_gzip', got '%s'", configs[0]["type"])
	}
}

func TestParserCustomSplit(t *testing.T) {
	parser := NewParser()
	sub := `split_string(separator="|")
send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Check separator
	if configs[0]["separator"] != "|" {
		t.Errorf("Expected separator '|', got '%v'", configs[0]["separator"])
	}
}

func TestParserAssignment(t *testing.T) {
	parser := NewParser()
	sub := `$.processed_data = decompress_gzip()
$.lines = split_string(source=$.processed_data, separator="\n")
$.output = send_stdout(source=$.lines)`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check that assignments have target field
	for i, config := range configs {
		if config["target"] == nil {
			t.Errorf("Config %d missing target field", i)
		}
	}
}

func TestParserComments(t *testing.T) {
	parser := NewParser()
	sub := `# This is a comment
split_string(separator="\n")
# Another comment
send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs (comments should be ignored), got %d", len(configs))
	}
}

func TestParserEmptyLines(t *testing.T) {
	parser := NewParser()
	sub := `split_string(separator="\n")

send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs (empty lines should be ignored), got %d", len(configs))
	}
}

func TestParserFunctionVariants(t *testing.T) {
	parser := NewParser()
	sub := `decompress_gzip()
split_string(separator="\n")
send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check function name variants
	if configs[0]["type"] != "decompress_gzip" {
		t.Errorf("Expected type 'decompress_gzip', got '%s'", configs[0]["type"])
	}
	if configs[1]["type"] != "split_string" {
		t.Errorf("Expected type 'split_string', got '%s'", configs[1]["type"])
	}
	if configs[2]["type"] != "send_stdout" {
		t.Errorf("Expected type 'send_stdout', got '%s'", configs[2]["type"])
	}
}

func TestParserArguments(t *testing.T) {
	parser := NewParser()
	sub := `split_string(separator="|", limit=10)
custom_function("arg1", "arg2", key="value", number=42, boolean=true)`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Check split arguments
	if configs[0]["separator"] != "|" {
		t.Errorf("Expected separator '|', got '%v'", configs[0]["separator"])
	}

	// Check custom function arguments
	if configs[1]["arg0"] != "arg1" {
		t.Errorf("Expected arg0 'arg1', got '%v'", configs[1]["arg0"])
	}
	if configs[1]["arg1"] != "arg2" {
		t.Errorf("Expected arg1 'arg2', got '%v'", configs[1]["arg1"])
	}
	if configs[1]["key"] != "value" {
		t.Errorf("Expected key 'value', got '%v'", configs[1]["key"])
	}
	if configs[1]["number"] != "42" {
		t.Errorf("Expected number '42', got '%v'", configs[1]["number"])
	}
	if configs[1]["boolean"] != "true" {
		t.Errorf("Expected boolean 'true', got '%v'", configs[1]["boolean"])
	}
}

func TestParserErrorCases(t *testing.T) {
	testCases := []struct {
		name string
		sub  string
	}{
		{"invalid function call", "invalid_function"},
		{"missing closing paren", "split_string(separator=\"\n\""},
		{"empty assignment", "foo ="},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()
			_, err := parser.Parse(tc.sub)
			if err == nil {
				t.Errorf("Expected error for '%s', but got none", tc.sub)
			}
		})
	}
}

func TestParserTargetedSplit(t *testing.T) {
	parser := NewParser()
	sub := `split_string(separator="|", source=$.foo)
send_stdout()`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	if configs[0]["type"] != "split_string" {
		t.Errorf("Expected type 'split_string', got '%s'", configs[0]["type"])
	}
	if configs[0]["separator"] != "|" {
		t.Errorf("Expected separator '|', got '%v'", configs[0]["separator"])
	}
	if configs[0]["source"] != "$.foo" {
		t.Errorf("Expected source '$.foo', got '%v'", configs[0]["source"])
	}
}

func TestParserDirectFieldAssignment(t *testing.T) {
	parser := NewParser()
	sub := `$.foo = $.message`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	config := configs[0]
	if config["type"] != "direct_assignment" {
		t.Errorf("Expected type 'direct_assignment', got '%s'", config["type"])
	}

	source, ok := config["source"].(string)
	if !ok {
		t.Fatal("Expected 'source' setting to be a string")
	}
	if source != "$.message" {
		t.Errorf("Expected source '$.message', got '%s'", source)
	}

	target, ok := config["target"].(string)
	if !ok {
		t.Fatal("Expected 'target' setting to be a string")
	}
	if target != "$.foo" {
		t.Errorf("Expected target '$.foo', got '%s'", target)
	}
}

func TestParserNestedFunctions(t *testing.T) {
	parser := NewParser()
	sub := `$.result = lowercase_string(source=decode_base64(source=$.encoded_data))
send_stdout(source=$.result)`

	configs, err := parser.Parse(sub)
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	// Should have 3 configs: decode_base64, lowercase_string, send_stdout
	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check first config (decode_base64)
	if configs[0]["type"] != "decode_base64" {
		t.Errorf("Expected type 'decode_base64', got '%s'", configs[0]["type"])
	}
	if configs[0]["source"] != "$.encoded_data" {
		t.Errorf("Expected source '$.encoded_data', got '%v'", configs[0]["source"])
	}

	// Check second config (lowercase_string)
	if configs[1]["type"] != "lowercase_string" {
		t.Errorf("Expected type 'lowercase_string', got '%s'", configs[1]["type"])
	}
	if configs[1]["source"] != "$.nested_output" {
		t.Errorf("Expected source '$.nested_output', got '%v'", configs[1]["source"])
	}

	// Check third config (send_stdout)
	if configs[2]["type"] != "send_stdout" {
		t.Errorf("Expected type 'send_stdout', got '%s'", configs[2]["type"])
	}
	if configs[2]["source"] != "$.result" {
		t.Errorf("Expected source '$.result', got '%v'", configs[2]["source"])
	}
}

func TestParserRejectsPositionalArgsForBuiltins(t *testing.T) {
	tests := []string{
		`split_string("|")`,
		`decompress_gzip("foo")`,
		`send_stdout("bar")`,
		`decode_base64("baz")`,
		`lowercase_string("qux")`,
	}
	for _, sub := range tests {
		parser := NewParser()
		_, err := parser.Parse(sub)
		if err == nil {
			t.Errorf("Expected error for positional args in built-in transform: %q", sub)
		}
	}
}

func TestParserAllowsValidFirstPositionalArgs(t *testing.T) {
	// These should work - first arg is a JSON path
	validTests := []string{
		`lowercase_string($.foo)`,
		`decode_base64($.bar)`,
		`send_stdout($.baz)`,
		`split_string($.qux)`,
		`decompress_gzip($.data)`,
	}
	for _, sub := range validTests {
		parser := NewParser()
		configs, err := parser.Parse(sub)
		if err != nil {
			t.Errorf("Expected success for valid first positional arg: %q, got error: %v", sub, err)
		}
		if len(configs) != 1 {
			t.Errorf("Expected 1 config for: %q, got %d", sub, len(configs))
		}
		if configs[0]["source"] == nil {
			t.Errorf("Expected source to be set for: %q", sub)
		}
	}

	// These should work - first arg is a function call
	validNestedTests := []string{
		`lowercase_string(decode_base64($.foo))`,
		`send_stdout(lowercase_string($.bar))`,
	}
	for _, sub := range validNestedTests {
		parser := NewParser()
		configs, err := parser.Parse(sub)
		if err != nil {
			t.Errorf("Expected success for valid nested function: %q, got error: %v", sub, err)
		}
		if len(configs) < 2 {
			t.Errorf("Expected at least 2 configs for nested function: %q, got %d", sub, len(configs))
		}
	}
}

func TestParserRejectsInvalidFirstPositionalArgs(t *testing.T) {
	// These should fail - first arg is neither a path nor a function
	invalidTests := []string{
		`split_string("\n", source="$.foo")`, // First arg is quoted string, second is named
		`split_string("\n")`,                 // First arg is quoted string
		`lowercase_string("hello")`,          // First arg is unquoted string
		`decode_base64(123)`,                 // First arg is number
		`send_stdout(true)`,                  // First arg is boolean
	}
	for _, sub := range invalidTests {
		parser := NewParser()
		_, err := parser.Parse(sub)
		if err == nil {
			t.Errorf("Expected error for invalid first positional arg: %q", sub)
		}
	}
}
