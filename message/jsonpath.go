package message

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// JSONPath represents a path to a value in a JSON object
type JSONPath struct {
	parts []string
}

// NewJSONPath creates a new JSONPath from a dot-separated string
func NewJSONPath(path string) *JSONPath {
	if path == "" {
		return &JSONPath{parts: []string{}}
	}
	return &JSONPath{parts: strings.Split(path, ".")}
}

// Get retrieves a value from a JSON object using the path
func (p *JSONPath) Get(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	return p.getFromInterface(obj)
}

// Set sets a value in a JSON object using the path
func (p *JSONPath) Set(data []byte, value interface{}) ([]byte, error) {
	if len(data) == 0 {
		data = []byte("{}")
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	obj, err := p.setInInterface(obj, value)
	if err != nil {
		return nil, err
	}

	return json.Marshal(obj)
}

// Delete removes a value from a JSON object using the path
func (p *JSONPath) Delete(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	obj, err := p.deleteFromInterface(obj)
	if err != nil {
		return nil, err
	}

	return json.Marshal(obj)
}

// getFromInterface recursively traverses the object to get the value
func (p *JSONPath) getFromInterface(obj interface{}) (interface{}, error) {
	if len(p.parts) == 0 {
		return obj, nil
	}

	current := obj
	for i, part := range p.parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, exists := v[part]; exists {
				current = val
			} else {
				return nil, fmt.Errorf("key '%s' not found at path '%s'", part, strings.Join(p.parts[:i+1], "."))
			}
		case []interface{}:
			// Handle array access like "0", "1", etc.
			if idx, err := strconv.Atoi(part); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil, fmt.Errorf("invalid array index '%s' at path '%s'", part, strings.Join(p.parts[:i+1], "."))
			}
		default:
			return nil, fmt.Errorf("cannot access key '%s' in non-object/non-array at path '%s'", part, strings.Join(p.parts[:i+1], "."))
		}
	}

	return current, nil
}

// setInInterface recursively traverses the object to set the value
func (p *JSONPath) setInInterface(obj interface{}, value interface{}) (interface{}, error) {
	if len(p.parts) == 0 {
		return value, nil
	}

	// If obj is nil, create a new map
	if obj == nil {
		obj = make(map[string]interface{})
	}

	// Handle root level
	if len(p.parts) == 1 {
		switch v := obj.(type) {
		case map[string]interface{}:
			v[p.parts[0]] = value
			return v, nil
		default:
			return nil, fmt.Errorf("cannot set key '%s' in non-object", p.parts[0])
		}
	}

	// Navigate to the parent of the target
	parentPath := &JSONPath{parts: p.parts[:len(p.parts)-1]}
	parent, err := parentPath.getFromInterface(obj)
	if err != nil {
		// If parent doesn't exist, create it
		parent = make(map[string]interface{})
		obj, err = parentPath.setInInterface(obj, parent)
		if err != nil {
			return nil, err
		}
		parent, _ = parentPath.getFromInterface(obj)
	}

	// Set the value in the parent
	switch v := parent.(type) {
	case map[string]interface{}:
		v[p.parts[len(p.parts)-1]] = value
	case []interface{}:
		if idx, err := strconv.Atoi(p.parts[len(p.parts)-1]); err == nil && idx >= 0 {
			// Extend array if necessary
			for len(v) <= idx {
				v = append(v, nil)
			}
			v[idx] = value
			// Update the parent in the original object
			obj, _ = parentPath.setInInterface(obj, v)
		} else {
			return nil, fmt.Errorf("invalid array index '%s'", p.parts[len(p.parts)-1])
		}
	default:
		return nil, fmt.Errorf("cannot set key '%s' in non-object/non-array", p.parts[len(p.parts)-1])
	}

	return obj, nil
}

// deleteFromInterface recursively traverses the object to delete the value
func (p *JSONPath) deleteFromInterface(obj interface{}) (interface{}, error) {
	if len(p.parts) == 0 {
		return obj, nil
	}

	// Navigate to the parent of the target
	if len(p.parts) == 1 {
		switch v := obj.(type) {
		case map[string]interface{}:
			delete(v, p.parts[0])
			return v, nil
		default:
			return nil, fmt.Errorf("cannot delete key '%s' from non-object", p.parts[0])
		}
	}

	parentPath := &JSONPath{parts: p.parts[:len(p.parts)-1]}
	parent, err := parentPath.getFromInterface(obj)
	if err != nil {
		// If parent doesn't exist, nothing to delete
		return obj, nil
	}

	// Delete the value from the parent
	switch v := parent.(type) {
	case map[string]interface{}:
		delete(v, p.parts[len(p.parts)-1])
		// Update the parent in the original object
		obj, _ = parentPath.setInInterface(obj, v)
	case []interface{}:
		if idx, err := strconv.Atoi(p.parts[len(p.parts)-1]); err == nil && idx >= 0 && idx < len(v) {
			// Set to nil instead of removing to maintain array structure
			v[idx] = nil
			// Update the parent in the original object
			obj, _ = parentPath.setInInterface(obj, v)
		}
	}

	return obj, nil
}
