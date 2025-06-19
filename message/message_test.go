package message

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func normalizeJSON(s string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return s
	}
	normalized, _ := json.Marshal(obj)
	return string(normalized)
}

func TestMessageMetadataPaths(t *testing.T) {
	msg := New()

	// Set up test data and metadata
	msg.SetData([]byte(`{"data_field": "data_value"}`))
	msg.SetMetadata([]byte(`{"meta_field": "meta_value", "nested": {"key": "value"}}`))

	// Test GetPathValue with metadata paths
	tests := []struct {
		path     string
		expected string
		exists   bool
	}{
		{"meta.$.meta_field", "meta_value", true},
		{"meta.$.nested.key", "value", true},
		{"meta.$.nonexistent", "", false},
		{"$.data_field", "data_value", true}, // Regular data path should still work
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			val := msg.GetPathValue(test.path)
			if val.Exists() != test.exists {
				t.Errorf("expected exists=%v, got %v", test.exists, val.Exists())
			}
			if test.exists && val.String() != test.expected {
				t.Errorf("expected %q, got %q", test.expected, val.String())
			}
		})
	}
}

func TestMessageSetMetadataPaths(t *testing.T) {
	msg := New()

	// Set up initial metadata
	msg.SetMetadata([]byte(`{"existing": "value"}`))

	// Test SetPathValue with metadata paths
	tests := []struct {
		path  string
		value interface{}
	}{
		{"meta.$.new_field", "new_value"},
		{"meta.$.nested.key", "nested_value"},
		{"meta.$.number", 42},
		{"meta.$.boolean", true},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			err := msg.SetPathValue(test.path, test.value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify the value was set correctly
			val := msg.GetPathValue(test.path)
			if !val.Exists() {
				t.Errorf("value was not set for path %s", test.path)
			}

			// Check the value matches
			switch expected := test.value.(type) {
			case string:
				if val.String() != expected {
					t.Errorf("expected %q, got %q", expected, val.String())
				}
			case int:
				if val.Int() != int64(expected) {
					t.Errorf("expected %d, got %d", expected, val.Int())
				}
			case bool:
				if val.Bool() != expected {
					t.Errorf("expected %v, got %v", expected, val.Bool())
				}
			}
		})
	}

	// Verify the final metadata structure
	expectedMeta := `{"existing":"value","new_field":"new_value","nested":{"key":"nested_value"},"number":42,"boolean":true}`
	actualMeta := string(msg.Metadata())
	if !jsonEqual(actualMeta, expectedMeta) {
		t.Errorf("expected metadata %s, got %s", expectedMeta, actualMeta)
	}
}

func TestMessageMetadataPathAccess(t *testing.T) {
	msg := New()

	// Set up test data with metadata
	msg.SetData([]byte(`{"name": "test", "value": 42}`))
	msg.SetMetadata([]byte(`{"source": "file", "timestamp": "2023-01-01"}`))

	tests := []struct {
		name     string
		path     string
		expected interface{}
		exists   bool
	}{
		{
			name:     "access metadata field with meta.$ syntax",
			path:     "meta.$.source",
			expected: "file",
			exists:   true,
		},
		{
			name:     "access metadata field with meta.$ syntax - nested",
			path:     "meta.$.timestamp",
			expected: "2023-01-01",
			exists:   true,
		},
		{
			name:     "access data field with $. syntax",
			path:     "$.name",
			expected: "test",
			exists:   true,
		},
		{
			name:     "access data field with $. syntax - nested",
			path:     "$.value",
			expected: float64(42),
			exists:   true,
		},
		{
			name:     "access non-existent metadata field",
			path:     "meta.$.nonexistent",
			expected: nil,
			exists:   false,
		},
		{
			name:     "access non-existent data field",
			path:     "$.nonexistent",
			expected: nil,
			exists:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := msg.GetPathValue(tt.path)
			if value.Exists() != tt.exists {
				t.Errorf("GetPathValue() exists = %v, want %v", value.Exists(), tt.exists)
			}
			if tt.exists && !reflect.DeepEqual(value.Value(), tt.expected) {
				t.Errorf("GetPathValue() value = %v, want %v", value.Value(), tt.expected)
			}
		})
	}
}

func TestMessageMetadataPathSet(t *testing.T) {
	msg := New()

	// Set up test data
	msg.SetData([]byte(`{"name": "test"}`))
	msg.SetMetadata([]byte(`{"source": "file"}`))

	tests := []struct {
		name     string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "set metadata field with meta.$ syntax",
			path:     "meta.$.timestamp",
			value:    "2023-01-01",
			expected: `{"source":"file","timestamp":"2023-01-01"}`,
		},
		{
			name:     "set data field with $. syntax",
			path:     "$.value",
			value:    42,
			expected: `{"name":"test","value":42}`,
		},
		{
			name:     "set nested metadata field",
			path:     "meta.$.nested.field",
			value:    "nested_value",
			expected: `{"source":"file","nested":{"field":"nested_value"}}`,
		},
		{
			name:     "set nested data field",
			path:     "$.nested.field",
			value:    "nested_value",
			expected: `{"name":"test","nested":{"field":"nested_value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh message for each test case
			msg := New()
			msg.SetData([]byte(`{"name": "test"}`))
			msg.SetMetadata([]byte(`{"source": "file"}`))

			err := msg.SetPathValue(tt.path, tt.value)
			if err != nil {
				t.Errorf("SetPathValue() error = %v", err)
				return
			}

			// Check the result
			if strings.HasPrefix(tt.path, "meta.$") {
				metadata := string(msg.Metadata())
				if !jsonEqual(metadata, tt.expected) {
					t.Errorf("SetPathValue() metadata = %s, want %s", normalizeJSON(metadata), normalizeJSON(tt.expected))
				}
			} else {
				data := string(msg.Data())
				if !jsonEqual(data, tt.expected) {
					t.Errorf("SetPathValue() data = %s, want %s", normalizeJSON(data), normalizeJSON(tt.expected))
				}
			}
		})
	}
}
