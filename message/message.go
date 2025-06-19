// Package message provides functions for managing data used by conditions and transforms.
package message

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	// metaKey is a prefix used to access the meta field in a Message.
	metaKey = "meta "
)

// errSetRawInvalidValue is returned when setRaw receives an invalid interface type.
var errSetRawInvalidValue = fmt.Errorf("invalid value type")

// Message is the data structure that is handled by transforms and interpreted by
// conditions.
//
// Data in each message can be accessed and modified as JSON text or binary data:
//   - JSON text is accessed using the GetValue, SetValue, and DeleteValue methods.
//   - Binary data is accessed using the Data and SetData methods.
//
// Metadata is an additional data field that is meant to store information about the
// message, but can be used for any purpose. For JSON text, metadata is accessed using
// the GetValue, SetValue, and DeleteValue methods with a key prefixed with "meta" (e.g.
// "meta foo"). Binary metadata is accessed using the Metadata and SetMetadata methods.
//
// Messages can also be configured as "control messages." Control messages are used for flow
// control in Substation functions and applications, but can be used for any purpose depending
// on the needs of a transform or condition. These messages should not contain data or metadata.
type Message struct {
	data []byte
	meta []byte

	// ctrl is a flag that indicates if the message is a control message.
	//
	// Control messages trigger special behavior in transforms and conditions.
	ctrl bool
}

// String returns the message data as a string.
func (m *Message) String() string {
	return string(m.data)
}

// New returns a new Message.
func New(opts ...func(*Message)) *Message {
	msg := &Message{}
	for _, o := range opts {
		o(msg)
	}

	return msg
}

// AsControl sets the message as a control message.
func (m *Message) AsControl() *Message {
	m.data = nil
	m.meta = nil

	m.ctrl = true
	return m
}

// IsControl returns true if the message is a control message.
func (m *Message) IsControl() bool {
	return m.ctrl
}

// Data returns the message data.
func (m *Message) Data() []byte {
	if m.ctrl {
		return nil
	}

	return m.data
}

// SetData sets the message data.
func (m *Message) SetData(data []byte) *Message {
	if m.ctrl {
		return m
	}

	m.data = data
	return m
}

// Metadata returns the message metadata.
func (m *Message) Metadata() []byte {
	if m.ctrl {
		return nil
	}

	return m.meta
}

// SetMetadata sets the message metadata.
func (m *Message) SetMetadata(metadata []byte) *Message {
	if m.ctrl {
		return m
	}

	m.meta = metadata
	return m
}

// isValidJSONPath returns true if the path is a valid JSONPath (starts with $. or meta.$.)
func isValidJSONPath(path string) bool {
	path = strings.TrimSpace(path)
	return path == "$" || path == "meta.$" || strings.HasPrefix(path, "$.") || strings.HasPrefix(path, "meta.$.")
}

// GetValue returns a value from the message data or metadata using a JSON path.
//
// The path must be a valid JSONPath:
// - Data: "$.foo", "$.nested.field"
// - Metadata: "meta.$.foo", "meta.$.nested.field"
//
// If the path is not valid, returns a non-existent value.
func (m *Message) GetValue(path string) Value {
	path = strings.TrimSpace(path)
	if !isValidJSONPath(path) {
		return Value{value: nil, exists: false}
	}

	if path == "$" {
		// Return the entire data object
		var obj interface{}
		if err := json.Unmarshal(m.data, &obj); err != nil {
			return Value{value: nil, exists: false}
		}
		return Value{value: obj, exists: true}
	}
	if path == "meta.$" {
		// Return the entire metadata object
		var obj interface{}
		if err := json.Unmarshal(m.meta, &obj); err != nil {
			return Value{value: nil, exists: false}
		}
		return Value{value: obj, exists: true}
	}

	if strings.HasPrefix(path, "meta.$.") {
		jsonPath := NewJSONPath(path)
		val, err := jsonPath.Get(m.meta)
		if err != nil {
			return Value{value: nil, exists: false}
		}
		return Value{value: val, exists: true}
	}

	if strings.HasPrefix(path, "$.") {
		jsonPath := NewJSONPath(path)
		val, err := jsonPath.Get(m.data)
		if err != nil {
			return Value{value: nil, exists: false}
		}
		return Value{value: val, exists: true}
	}

	return Value{value: nil, exists: false}
}

// SetValue sets a value in the message data or metadata using a JSON path.
//
// The path must be a valid JSONPath:
// - Data: "$.foo", "$.nested.field"
// - Metadata: "meta.$.foo", "meta.$.nested.field"
//
// If the path is not valid, returns an error.
func (m *Message) SetValue(path string, value interface{}) error {
	path = strings.TrimSpace(path)
	if !isValidJSONPath(path) {
		return fmt.Errorf("invalid JSONPath: %s", path)
	}

	if path == "$" {
		// Set the entire data object
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		m.data = data
		return nil
	}
	if path == "meta.$" {
		// Set the entire metadata object
		meta, err := json.Marshal(value)
		if err != nil {
			return err
		}
		m.meta = meta
		return nil
	}

	if strings.HasPrefix(path, "meta.$.") {
		jsonPath := NewJSONPath(path)
		meta, err := jsonPath.Set(m.meta, value)
		if err != nil {
			return err
		}
		m.meta = meta
		return nil
	}

	if strings.HasPrefix(path, "$.") {
		jsonPath := NewJSONPath(path)
		data, err := jsonPath.Set(m.data, value)
		if err != nil {
			return err
		}
		m.data = data
		return nil
	}

	return fmt.Errorf("invalid JSONPath: %s", path)
}

// DeleteValue deletes a value in the message data or metadata using a JSON path.
//
// The path must be a valid JSONPath:
// - Data: "$.foo", "$.nested.field"
// - Metadata: "meta.$.foo", "meta.$.nested.field"
//
// If the path is not valid, returns an error.
func (m *Message) DeleteValue(path string) error {
	path = strings.TrimSpace(path)
	if !isValidJSONPath(path) {
		return fmt.Errorf("invalid JSONPath: %s", path)
	}

	if path == "$" {
		// Delete entire data object
		m.data = []byte(`{}`)
		return nil
	}
	if path == "meta.$" {
		// Delete entire metadata object
		m.meta = []byte(`{}`)
		return nil
	}

	if strings.HasPrefix(path, "meta.$.") {
		jsonPath := NewJSONPath(path)
		meta, err := jsonPath.Delete(m.meta)
		if err != nil {
			return err
		}
		m.meta = meta
		return nil
	}

	if strings.HasPrefix(path, "$.") {
		jsonPath := NewJSONPath(path)
		data, err := jsonPath.Delete(m.data)
		if err != nil {
			return err
		}
		m.data = data
		return nil
	}

	return fmt.Errorf("invalid JSONPath: %s", path)
}

// Value provides access to JSON values returned by GetValue.
type Value struct {
	value  interface{}
	exists bool
}

// Value returns the underlying value.
func (v Value) Value() any {
	return v.value
}

// String returns the value as a string.
func (v Value) String() string {
	if v.value == nil {
		return ""
	}
	switch s := v.value.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		// Try to marshal to JSON string
		if b, err := json.Marshal(v.value); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v.value)
	}
}

// Bytes returns the value as bytes.
func (v Value) Bytes() []byte {
	if v.value == nil {
		return nil
	}
	switch b := v.value.(type) {
	case []byte:
		return b
	case string:
		return []byte(b)
	default:
		// Try to marshal to JSON
		if jsonBytes, err := json.Marshal(v.value); err == nil {
			return jsonBytes
		}
		return []byte(fmt.Sprintf("%v", v.value))
	}
}

// Int returns the value as an int64.
func (v Value) Int() int64 {
	if v.value == nil {
		return 0
	}
	switch n := v.value.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	case string:
		if i, err := strconv.ParseInt(n, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// Uint returns the value as a uint64.
func (v Value) Uint() uint64 {
	if v.value == nil {
		return 0
	}
	switch n := v.value.(type) {
	case uint:
		return uint64(n)
	case uint64:
		return n
	case int:
		if n >= 0 {
			return uint64(n)
		}
	case int64:
		if n >= 0 {
			return uint64(n)
		}
	case float64:
		if n >= 0 {
			return uint64(n)
		}
	case string:
		if u, err := strconv.ParseUint(n, 10, 64); err == nil {
			return u
		}
	}
	return 0
}

// Float returns the value as a float64.
func (v Value) Float() float64 {
	if v.value == nil {
		return 0
	}
	switch n := v.value.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f
		}
	}
	return 0
}

// Bool returns the value as a bool.
func (v Value) Bool() bool {
	if v.value == nil {
		return false
	}
	switch b := v.value.(type) {
	case bool:
		return b
	case string:
		return b == "true" || b == "1"
	case int:
		return b != 0
	case float64:
		return b != 0
	}
	return false
}

// Array returns the value as an array of Values.
func (v Value) Array() []Value {
	if v.value == nil {
		return nil
	}
	switch arr := v.value.(type) {
	case []interface{}:
		result := make([]Value, len(arr))
		for i, item := range arr {
			result[i] = Value{value: item, exists: true}
		}
		return result
	case []Value:
		return arr
	}
	return nil
}

// IsArray returns true if the value is an array.
func (v Value) IsArray() bool {
	if v.value == nil {
		return false
	}
	switch v.value.(type) {
	case []interface{}, []Value:
		return true
	}
	return false
}

// Map returns the value as a map of Values.
func (v Value) Map() map[string]Value {
	if v.value == nil {
		return nil
	}
	switch m := v.value.(type) {
	case map[string]interface{}:
		result := make(map[string]Value)
		for k, v := range m {
			result[k] = Value{value: v, exists: true}
		}
		return result
	case map[string]Value:
		return m
	}
	return nil
}

// Exists returns true if the value exists.
func (v Value) Exists() bool {
	return v.exists && v.value != nil
}

func deleteValue(json []byte, key string) ([]byte, error) {
	if len(json) == 0 {
		return json, nil
	}

	if !utf8.Valid(json) {
		return json, nil
	}

	path := NewJSONPath(key)
	return path.Delete(json)
}

func setValue(obj []byte, key string, value interface{}) ([]byte, error) {
	if len(obj) == 0 {
		obj = []byte("{}")
	}

	if !utf8.Valid(obj) {
		return obj, nil
	}

	if !validJSON(value) {
		return obj, errSetRawInvalidValue
	}

	path := NewJSONPath(key)
	return path.Set(obj, value)
}

func setRaw(json []byte, key string, value interface{}) ([]byte, error) {
	if len(json) == 0 {
		json = []byte("{}")
	}

	if !utf8.Valid(json) {
		return json, nil
	}

	result, err := setValue(json, key, anyToBytes(value))
	if err != nil {
		return json, err
	}

	return result, nil
}

func validJSON(data interface{}) bool {
	switch data.(type) {
	case string, bool, float64, int, int64, uint, uint64, nil:
		return true
	case []interface{}, map[string]interface{}:
		return true
	default:
		return false
	}
}

func anyToBytes(v any) []byte {
	msg := New()
	_ = msg.SetValue("_", v)

	return msg.GetValue("_").Bytes()
}
