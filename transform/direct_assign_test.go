package transform

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jshlbrd/vibestation/message"
)

func TestDirectAssign_Basic(t *testing.T) {
	msgData := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
	}

	jsonData, err := json.Marshal(msgData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	msg := message.New()
	msg.SetData(jsonData)

	// Create transformer that deletes $.field2
	transformer := newDirectDeleteTransformer("$.field2")

	// Transform the message
	result, err := transformer.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(result))
	}

	resultMsg := result[0]

	// Check that field2 was deleted
	deletedValue := resultMsg.GetValue("$.field2")
	if deletedValue.Exists() {
		t.Fatal("Expected $.field2 to be deleted")
	}

	// Check that the deleted value is stored in $.deleted_value
	storedValue := resultMsg.GetValue("$.deleted_value")
	if !storedValue.Exists() {
		t.Fatal("Expected $.deleted_value to exist")
	}
	if storedValue.Value() != "value2" {
		t.Errorf("Expected deleted_value 'value2', got %v", storedValue.Value())
	}

	// Verify other fields still exist
	field1Value := resultMsg.GetValue("$.field1")
	if !field1Value.Exists() {
		t.Fatal("Expected $.field1 to still exist")
	}
	if field1Value.Value() != "value1" {
		t.Errorf("Expected field1 'value1', got %v", field1Value.Value())
	}

	field3Value := resultMsg.GetValue("$.field3")
	if !field3Value.Exists() {
		t.Fatal("Expected $.field3 to still exist")
	}
	if field3Value.Value() != "value3" {
		t.Errorf("Expected field3 'value3', got %v", field3Value.Value())
	}
}

func TestDirectAssign_NonExistentField(t *testing.T) {
	msgData := map[string]interface{}{
		"field1": "value1",
	}

	jsonData, err := json.Marshal(msgData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	msg := message.New()
	msg.SetData(jsonData)

	// Create transformer that tries to delete non-existent field
	transformer := newDirectDeleteTransformer("$.non_existent")

	// Transform the message
	result, err := transformer.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(result))
	}

	resultMsg := result[0]

	// Check that existing field still exists
	field1Value := resultMsg.GetValue("$.field1")
	if !field1Value.Exists() {
		t.Fatal("Expected $.field1 to still exist")
	}

	// Check that deleted_value is not set for non-existent field
	deletedValue := resultMsg.GetValue("$.deleted_value")
	if deletedValue.Exists() {
		t.Fatal("Expected $.deleted_value to not exist for non-existent field")
	}
}

func TestDirectAssign_NestedField(t *testing.T) {
	msgData := map[string]interface{}{
		"nested": map[string]interface{}{
			"inner": "value",
		},
		"other": "other_value",
	}

	jsonData, err := json.Marshal(msgData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	msg := message.New()
	msg.SetData(jsonData)

	// Create transformer that deletes nested field
	transformer := newDirectDeleteTransformer("$.nested.inner")

	// Transform the message
	result, err := transformer.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(result))
	}

	resultMsg := result[0]

	// Check that nested field was deleted
	deletedValue := resultMsg.GetValue("$.nested.inner")
	if deletedValue.Exists() {
		t.Fatal("Expected $.nested.inner to be deleted")
	}

	// Check that the deleted value is stored
	storedValue := resultMsg.GetValue("$.deleted_value")
	if !storedValue.Exists() {
		t.Fatal("Expected $.deleted_value to exist")
	}
	if storedValue.Value() != "value" {
		t.Errorf("Expected deleted_value 'value', got %v", storedValue.Value())
	}

	// Verify nested object still exists but is empty
	nestedValue := resultMsg.GetValue("$.nested")
	if !nestedValue.Exists() {
		t.Fatal("Expected $.nested to still exist")
	}

	// Verify other field still exists
	otherValue := resultMsg.GetValue("$.other")
	if !otherValue.Exists() {
		t.Fatal("Expected $.other to still exist")
	}
}
