package message

import (
	"encoding/json"
	"testing"
)

func normalizeJSON(s string) string {
	var v interface{}
	json.Unmarshal([]byte(s), &v)
	b, _ := json.Marshal(v)
	return string(b)
}

func TestMessageGetValue(t *testing.T) {
	msg := New()

	// Set up test data and metadata
	msg.SetData([]byte(`{"data_field": "data_value"}`))
	msg.SetMetadata([]byte(`{"meta_field": "meta_value", "nested": {"key": "value"}}`))

	// Test GetValue with metadata paths
	metadataTests := []struct {
		path     string
		expected interface{}
		exists   bool
	}{
		{"meta.$.meta_field", "meta_value", true},
		{"meta.$.nested.key", "value", true},
		{"meta.$.nonexistent", nil, false},
		{"$.data_field", "data_value", true}, // Regular data path should still work
	}

	for _, test := range metadataTests {
		val := msg.GetValue(test.path)
		if val.Exists() != test.exists {
			t.Errorf("GetValue() exists = %v, want %v", val.Exists(), test.exists)
		}
		if val.Exists() && val.Value() != test.expected {
			t.Errorf("GetValue() value = %v, want %v", val.Value(), test.expected)
		}
	}

	// Test SetValue with metadata paths
	for _, test := range metadataTests {
		err := msg.SetValue(test.path, test.expected)
		if err != nil {
			t.Errorf("SetValue() error = %v", err)
		}
		val := msg.GetValue(test.path)
		if val.Exists() != test.exists {
			t.Errorf("SetValue() exists = %v, want %v", val.Exists(), test.exists)
		}
		if val.Exists() && val.Value() != test.expected {
			t.Errorf("SetValue() value = %v, want %v", val.Value(), test.expected)
		}
	}
}

func TestMessageSetValue(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		value    interface{}
		expected string
		isMeta   bool
	}{
		{
			name:     "set metadata field with meta.$ syntax",
			path:     "meta.$.timestamp",
			value:    "2023-01-01",
			expected: `{"source":"file","timestamp":"2023-01-01"}`,
			isMeta:   true,
		},
		{
			name:     "set data field with $. syntax",
			path:     "$.value",
			value:    42,
			expected: `{"name":"test","value":42}`,
			isMeta:   false,
		},
		{
			name:     "set nested metadata field",
			path:     "meta.$.nested.field",
			value:    "nested_value",
			expected: `{"source":"file","nested":{"field":"nested_value"}}`,
			isMeta:   true,
		},
		{
			name:     "set nested data field",
			path:     "$.nested.field",
			value:    "nested_value",
			expected: `{"name":"test","nested":{"field":"nested_value"}}`,
			isMeta:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh message for each test case
			msg := New()
			msg.SetData([]byte(`{"name": "test"}`))
			msg.SetMetadata([]byte(`{"source": "file"}`))

			err := msg.SetValue(tt.path, tt.value)
			if err != nil {
				t.Errorf("SetValue() error = %v", err)
			}

			if tt.isMeta {
				metadata := string(msg.Metadata())
				if normalizeJSON(metadata) != normalizeJSON(tt.expected) {
					t.Errorf("SetValue() metadata = %s, want %s", normalizeJSON(metadata), normalizeJSON(tt.expected))
				}
			} else {
				data := string(msg.Data())
				if normalizeJSON(data) != normalizeJSON(tt.expected) {
					t.Errorf("SetValue() data = %s, want %s", normalizeJSON(data), normalizeJSON(tt.expected))
				}
			}
		})
	}
}

func TestMessageInvalidJSONPath(t *testing.T) {
	msg := New()
	msg.SetData([]byte(`{"foo": "bar"}`))
	msg.SetMetadata([]byte(`{"meta": "data"}`))

	invalidPaths := []string{
		"foo",      // simple key
		"bar.baz",  // dot notation, no prefix
		"meta foo", // old meta key
		"",         // empty
		"$foo",     // missing dot
		"meta$foo", // missing dot
	}

	for _, path := range invalidPaths {
		t.Run("GetValue:"+path, func(t *testing.T) {
			val := msg.GetValue(path)
			if val.Exists() {
				t.Errorf("GetValue(%q) should not exist, got: %v", path, val.Value())
			}
		})
		t.Run("SetValue:"+path, func(t *testing.T) {
			err := msg.SetValue(path, "value")
			if err == nil {
				t.Errorf("SetValue(%q) should return error for invalid JSONPath", path)
			}
		})
		t.Run("DeleteValue:"+path, func(t *testing.T) {
			err := msg.DeleteValue(path)
			if err == nil {
				t.Errorf("DeleteValue(%q) should return error for invalid JSONPath", path)
			}
		})
	}
}

func TestMessageRootAndMetaRoot(t *testing.T) {
	msg := New()
	msg.SetData([]byte(`{"foo": 123, "bar": "baz"}`))
	msg.SetMetadata([]byte(`{"meta": true, "count": 5}`))

	// GetValue for $ and meta.$
	val := msg.GetValue("$")
	if !val.Exists() {
		t.Error("GetValue($) should exist")
	}
	obj, ok := val.Value().(map[string]interface{})
	if !ok || obj["foo"] != float64(123) || obj["bar"] != "baz" {
		t.Errorf("GetValue($) returned wrong object: %v", val.Value())
	}

	metaVal := msg.GetValue("meta.$")
	if !metaVal.Exists() {
		t.Error("GetValue(meta.$) should exist")
	}
	metaObj, ok := metaVal.Value().(map[string]interface{})
	if !ok || metaObj["meta"] != true || metaObj["count"] != float64(5) {
		t.Errorf("GetValue(meta.$) returned wrong object: %v", metaVal.Value())
	}

	// SetValue for $ and meta.$
	newData := map[string]interface{}{"x": 1, "y": 2}
	err := msg.SetValue("$", newData)
	if err != nil {
		t.Errorf("SetValue($) error: %v", err)
	}
	val = msg.GetValue("$")
	obj, ok = val.Value().(map[string]interface{})
	if !ok || obj["x"] != float64(1) || obj["y"] != float64(2) {
		t.Errorf("SetValue($) did not update data: %v", val.Value())
	}

	newMeta := map[string]interface{}{"meta": false, "z": 99}
	err = msg.SetValue("meta.$", newMeta)
	if err != nil {
		t.Errorf("SetValue(meta.$) error: %v", err)
	}
	metaVal = msg.GetValue("meta.$")
	metaObj, ok = metaVal.Value().(map[string]interface{})
	if !ok || metaObj["meta"] != false || metaObj["z"] != float64(99) {
		t.Errorf("SetValue(meta.$) did not update metadata: %v", metaVal.Value())
	}

	// DeleteValue for $ and meta.$
	err = msg.DeleteValue("$")
	if err != nil {
		t.Errorf("DeleteValue($) error: %v", err)
	}
	val = msg.GetValue("$")
	obj, ok = val.Value().(map[string]interface{})
	if !ok || len(obj) != 0 {
		t.Errorf("DeleteValue($) did not clear data: %v", val.Value())
	}
	err = msg.DeleteValue("meta.$")
	if err != nil {
		t.Errorf("DeleteValue(meta.$) error: %v", err)
	}
	metaVal = msg.GetValue("meta.$")
	metaObj, ok = metaVal.Value().(map[string]interface{})
	if !ok || len(metaObj) != 0 {
		t.Errorf("DeleteValue(meta.$) did not clear metadata: %v", metaVal.Value())
	}
}

func TestMessageDebug(t *testing.T) {
	msg := New()
	msg.SetData([]byte(`{"data_field": "data_value"}`))
	msg.SetMetadata([]byte(`{"meta_field": "meta_value", "nested": {"key": "value"}}`))

	// Test a simple metadata path
	val := msg.GetValue("meta.$.meta_field")
	t.Logf("GetValue('meta.$.meta_field') = %v, exists = %v", val.Value(), val.Exists())

	// Test a simple data path
	val = msg.GetValue("$.data_field")
	t.Logf("GetValue('$.data_field') = %v, exists = %v", val.Value(), val.Exists())
}
