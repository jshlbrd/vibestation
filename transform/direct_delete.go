package transform

import (
	"context"
	"fmt"

	"github.com/jshlbrd/vibestation/message"
)

// DirectDeleteTransformer removes a field from the message and returns its value
type DirectDeleteTransformer struct {
	path   string
	target string // If set, this is an assignment context
}

// newDirectDeleteTransformer creates a new direct delete transformer
func newDirectDeleteTransformer(path string) *DirectDeleteTransformer {
	return &DirectDeleteTransformer{
		path: path,
	}
}

// newDirectDeleteTransformerWithTarget creates a new direct delete transformer for assignment context
func newDirectDeleteTransformerWithTarget(path, target string) *DirectDeleteTransformer {
	return &DirectDeleteTransformer{
		path:   path,
		target: target,
	}
}

// Transform removes the specified field from the message and returns its value
func (d *DirectDeleteTransformer) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	// Get the value before deleting it (should be strict JSONPath)
	value := msg.GetValue(d.path)
	if !value.Exists() {
		// If field doesn't exist, return nil value
		return []*message.Message{msg}, nil
	}

	// Store the value to return
	deletedValue := value.Value()

	// Delete the field (should be strict JSONPath)
	err := msg.DeleteValue(d.path)
	if err != nil {
		return nil, fmt.Errorf("direct delete: failed to delete path %s: %v", d.path, err)
	}

	// If this is an assignment context, set the target (should be strict JSONPath)
	if d.target != "" {
		err = msg.SetValue(d.target, deletedValue)
		if err != nil {
			return nil, fmt.Errorf("direct delete: failed to set target %s: %v", d.target, err)
		}
	} else {
		// Set the deleted value in a special field for retrieval
		err = msg.SetValue("$.deleted_value", deletedValue)
		if err != nil {
			return nil, fmt.Errorf("delete: failed to store deleted value: %v", err)
		}
	}

	return []*message.Message{msg}, nil
}
