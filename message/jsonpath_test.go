package message

import (
	"encoding/json"
	"testing"
)

func TestJSONPath_Get(t *testing.T) {
	data := []byte(`{
		"a": {
			"b": {
				"c": "value"
			},
			"d": [1, 2, 3]
		},
		"e": "simple"
	}`)

	tests := []struct {
		path     string
		expected interface{}
		exists   bool
	}{
		{"$.a.b.c", "value", true},
		{"$.a.d[0]", float64(1), true},
		{"$.a.d[1]", float64(2), true},
		{"$.e", "simple", true},
		{"$.a.b.d", nil, false},
		{"$.x.y.z", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			path := NewJSONPath(tt.path)
			result, err := path.Get(data)

			if tt.exists {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for non-existent path")
				}
			}
		})
	}

	// Test root path separately
	t.Run("root path", func(t *testing.T) {
		path := NewJSONPath("")
		result, err := path.Get(data)
		if err != nil {
			t.Errorf("Expected no error for root path, got %v", err)
		}
		if result == nil {
			t.Error("Expected non-nil result for root path")
		}
	})
}

func TestJSONPath_Set(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		path     string
		value    interface{}
		expected string
	}{
		{
			name:     "set simple key",
			initial:  `{"a": "old"}`,
			path:     "$.a",
			value:    "new",
			expected: `{"a":"new"}`,
		},
		{
			name:     "set nested key",
			initial:  `{"a": {"b": "old"}}`,
			path:     "$.a.b",
			value:    "new",
			expected: `{"a":{"b":"new"}}`,
		},
		{
			name:     "create nested path",
			initial:  `{}`,
			path:     "$.a.b.c",
			value:    "value",
			expected: `{"a":{"b":{"c":"value"}}}`,
		},
		{
			name:     "set array element",
			initial:  `{"arr": [1, 2, 3]}`,
			path:     "$.arr[1]",
			value:    "new",
			expected: `{"arr":[1,"new",3]}`,
		},
		{
			name:     "set integer",
			initial:  `{}`,
			path:     "$.intVal",
			value:    42,
			expected: `{"intVal":42}`,
		},
		{
			name:     "set float",
			initial:  `{}`,
			path:     "$.floatVal",
			value:    3.14,
			expected: `{"floatVal":3.14}`,
		},
		{
			name:     "set array",
			initial:  `{}`,
			path:     "$.arrVal",
			value:    []interface{}{1, 2, 3},
			expected: `{"arrVal":[1,2,3]}`,
		},
		{
			name:     "set object",
			initial:  `{}`,
			path:     "$.objVal",
			value:    map[string]interface{}{"foo": 1, "bar": 2},
			expected: `{"objVal":{"bar":2,"foo":1}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewJSONPath(tt.path)
			result, err := path.Set([]byte(tt.initial), tt.value)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Normalize JSON for comparison
			var resultObj, expectedObj interface{}
			json.Unmarshal(result, &resultObj)
			json.Unmarshal([]byte(tt.expected), &expectedObj)

			if !jsonEqual(mustMarshal(resultObj), mustMarshal(expectedObj)) {
				t.Errorf("Expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}

func TestJSONPath_Delete(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		path     string
		expected string
	}{
		{
			name:     "delete simple key",
			initial:  `{"a": "value", "b": "keep"}`,
			path:     "$.a",
			expected: `{"b":"keep"}`,
		},
		{
			name:     "delete nested key",
			initial:  `{"a": {"b": "value", "c": "keep"}}`,
			path:     "$.a.b",
			expected: `{"a":{"c":"keep"}}`,
		},
		{
			name:     "delete non-existent key",
			initial:  `{"a": "value"}`,
			path:     "$.b",
			expected: `{"a":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := NewJSONPath(tt.path)
			result, err := path.Delete([]byte(tt.initial))
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			// Normalize JSON for comparison
			var resultObj, expectedObj interface{}
			json.Unmarshal(result, &resultObj)
			json.Unmarshal([]byte(tt.expected), &expectedObj)

			if !jsonEqual(mustMarshal(resultObj), mustMarshal(expectedObj)) {
				t.Errorf("Expected %s, got %s", tt.expected, string(result))
			}
		})
	}
}

func TestMessage_GetValue_SetValue(t *testing.T) {
	msg := New()

	// Test setting and getting nested values
	err := msg.SetValue("$.a.b.c", "nested_value")
	if err != nil {
		t.Errorf("Expected no error setting nested value, got %v", err)
	}

	val := msg.GetValue("$.a.b.c")
	if !val.Exists() {
		t.Error("Expected value to exist")
	}
	if val.String() != "nested_value" {
		t.Errorf("Expected 'nested_value', got '%s'", val.String())
	}

	// Test setting and getting array elements
	err = msg.SetValue("$.arr.0", "first")
	if err != nil {
		t.Errorf("Expected no error setting array element, got %v", err)
	}

	val = msg.GetValue("$.arr.0")
	if !val.Exists() {
		t.Error("Expected array element to exist")
	}
	if val.String() != "first" {
		t.Errorf("Expected 'first', got '%s'", val.String())
	}

	// Test metadata access
	err = msg.SetValue("meta.$.test", "metadata_value")
	if err != nil {
		t.Errorf("Expected no error setting metadata, got %v", err)
	}

	val = msg.GetValue("meta.$.test")
	if !val.Exists() {
		t.Error("Expected metadata to exist")
	}
	if val.String() != "metadata_value" {
		t.Errorf("Expected 'metadata_value', got '%s'", val.String())
	}
}

// Helper for marshaling to string
func mustMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
