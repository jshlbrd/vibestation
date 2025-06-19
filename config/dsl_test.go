package config

import (
	"testing"
)

func TestSUBParserBasic(t *testing.T) {
	sub := `split("\n")
print()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Check first config (split)
	if configs[0].Type != "string_split" {
		t.Errorf("Expected type 'string_split', got '%s'", configs[0].Type)
	}
	if configs[0].Settings["separator"] != "\n" {
		sep, _ := configs[0].Settings["separator"].(string)
		t.Errorf("Expected separator (newline, byte 10), got '%v' (bytes: %v)", sep, []byte(sep))
	}

	// Check second config (print)
	if configs[1].Type != "send_stdout" {
		t.Errorf("Expected type 'send_stdout', got '%s'", configs[1].Type)
	}
}

func TestSUBParserGzip(t *testing.T) {
	sub := `gzip_decompress()
split("\n")
print()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check first config (gzip_decompress)
	if configs[0].Type != "format_from_gzip" {
		t.Errorf("Expected type 'format_from_gzip', got '%s'", configs[0].Type)
	}
}

func TestSUBParserCustomSplit(t *testing.T) {
	sub := `split("|")
print()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Check separator
	if configs[0].Settings["separator"] != "|" {
		t.Errorf("Expected separator '|', got '%v'", configs[0].Settings["separator"])
	}
}

func TestSUBParserAssignment(t *testing.T) {
	sub := `$.processed_data = gzip_decompress()
$.lines = split($.processed_data, "\n")
$.output = print($.lines)`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check that assignments have target field
	for i, config := range configs {
		if config.Settings["target"] == nil {
			t.Errorf("Config %d missing target field", i)
		}
	}
}

func TestSUBParserComments(t *testing.T) {
	sub := `# This is a comment
split("\n")
# Another comment
print()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs (comments should be ignored), got %d", len(configs))
	}
}

func TestSUBParserEmptyLines(t *testing.T) {
	sub := `split("\n")

print()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs (empty lines should be ignored), got %d", len(configs))
	}
}

func TestSUBParserFunctionVariants(t *testing.T) {
	sub := `decompress_gzip()
string_split("\n")
stdout()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check function name variants
	if configs[0].Type != "format_from_gzip" {
		t.Errorf("Expected type 'format_from_gzip', got '%s'", configs[0].Type)
	}
	if configs[1].Type != "string_split" {
		t.Errorf("Expected type 'string_split', got '%s'", configs[1].Type)
	}
	if configs[2].Type != "send_stdout" {
		t.Errorf("Expected type 'send_stdout', got '%s'", configs[2].Type)
	}
}

func TestSUBParserArguments(t *testing.T) {
	sub := `split("|", limit: 10)
custom_function("arg1", "arg2", key: "value", number: 42, boolean: true)`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Check split arguments
	if configs[0].Settings["separator"] != "|" {
		t.Errorf("Expected separator '|', got '%v'", configs[0].Settings["separator"])
	}

	// Check custom function arguments
	if configs[1].Settings["arg0"] != "arg1" {
		t.Errorf("Expected arg0 'arg1', got '%v'", configs[1].Settings["arg0"])
	}
	if configs[1].Settings["arg1"] != "arg2" {
		t.Errorf("Expected arg1 'arg2', got '%v'", configs[1].Settings["arg1"])
	}
	if configs[1].Settings["key"] != "value" {
		t.Errorf("Expected key 'value', got '%v'", configs[1].Settings["key"])
	}
	if configs[1].Settings["number"] != 42 {
		t.Errorf("Expected number 42, got '%v'", configs[1].Settings["number"])
	}
	if configs[1].Settings["boolean"] != true {
		t.Errorf("Expected boolean true, got '%v'", configs[1].Settings["boolean"])
	}
}

func TestSUBParserErrorCases(t *testing.T) {
	testCases := []struct {
		name string
		sub  string
	}{
		{"invalid function call", "invalid_function"},
		{"missing closing paren", "split(\"\n\""},
		{"empty assignment", "foo ="},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewSUBParser(tc.sub)
			_, err := parser.Parse()
			if err == nil {
				t.Errorf("Expected error for '%s', but got none", tc.sub)
			}
		})
	}
}

func TestSUBParserTargetedSplit(t *testing.T) {
	sub := `split(separator="|", input=$.foo)
print()`

	parser := NewSUBParser(sub)
	configs, err := parser.Parse()
	if err != nil {
		t.Fatalf("Failed to parse SUB: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	if configs[0].Type != "string_split" {
		t.Errorf("Expected type 'string_split', got '%s'", configs[0].Type)
	}
	if configs[0].Settings["separator"] != "|" {
		t.Errorf("Expected separator '|', got '%v'", configs[0].Settings["separator"])
	}
	if configs[0].Settings["input"] != "$.foo" {
		t.Errorf("Expected input '$.foo', got '%v'", configs[0].Settings["input"])
	}
}
