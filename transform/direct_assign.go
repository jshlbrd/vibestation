package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jshlbrd/vibestation/message"
)

// DirectAssignTransformer handles direct field assignments like "$.foo = $.bar"
type DirectAssignTransformer struct {
	source string
	target string
}

// newDirectAssignTransformer creates a new direct assign transformer
func newDirectAssignTransformer(source, target string) *DirectAssignTransformer {
	return &DirectAssignTransformer{
		source: source,
		target: target,
	}
}

// Transform copies a value from source path to target path
func (d *DirectAssignTransformer) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	var value interface{}

	// Handle special case: source is "$" (entire message object)
	if d.source == "$" {
		// Get the entire message data as parsed JSON
		data := msg.Data()
		if len(data) == 0 {
			// If no data, skip the assignment
			return []*message.Message{msg}, nil
		}

		// Parse the JSON data to get the entire object
		var parsedData interface{}
		if err := json.Unmarshal(data, &parsedData); err != nil {
			return nil, fmt.Errorf("direct assign: failed to parse message data as JSON: %v", err)
		}
		value = parsedData
	} else {
		// Get the value from source path (should be strict JSONPath)
		sourceValue := msg.GetValue(d.source)
		if !sourceValue.Exists() {
			// If source doesn't exist, skip the assignment
			return []*message.Message{msg}, nil
		}
		value = sourceValue.Value()
	}

	// Set the value to target path (should be strict JSONPath)
	err := msg.SetValue(d.target, value)
	if err != nil {
		return nil, fmt.Errorf("direct assign: failed to set target %s: %v", d.target, err)
	}

	return []*message.Message{msg}, nil
}
