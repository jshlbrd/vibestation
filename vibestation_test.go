package vibestation

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

func jsonEqual(a, b string) bool {
	var o1, o2 interface{}
	if err := json.Unmarshal([]byte(a), &o1); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &o2); err != nil {
		return false
	}
	return reflect.DeepEqual(o1, o2)
}

func TestVibestationTransform(t *testing.T) {
	// Create a simple configuration with string split and stdout transforms
	cfg := Config{
		Transforms: []config.Config{
			{
				Type: "string_split",
				Settings: map[string]interface{}{
					"separator": "\n",
					"id":        "test_split",
				},
			},
			{
				Type: "send_stdout",
				Settings: map[string]interface{}{
					"id": "test_stdout",
				},
			},
		},
	}

	// Create vibestation instance
	ctx := context.Background()
	vibe, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create vibestation: %v", err)
	}

	// Create test message
	testData := "line1\nline2\nline3"
	msg := message.New().SetData([]byte(testData))

	// Process the message
	results, err := vibe.Transform(ctx, msg)
	if err != nil {
		t.Fatalf("Failed to transform message: %v", err)
	}

	// Verify we got 3 messages (one for each line)
	if len(results) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(results))
	}

	// Verify the content of each message
	expected := []string{"line1", "line2", "line3"}
	for i, result := range results {
		if string(result.Data()) != expected[i] {
			t.Errorf("Message %d: expected '%s', got '%s'", i, expected[i], string(result.Data()))
		}
	}
}

func TestVibestationNoTransforms(t *testing.T) {
	// Test that vibestation returns an error when no transforms are configured
	cfg := Config{
		Transforms: nil,
	}

	ctx := context.Background()
	_, err := New(ctx, cfg)
	if err == nil {
		t.Error("Expected error when no transforms are configured")
	}
}

func TestVibestationDirectAssignment(t *testing.T) {
	// Create a config with direct field assignment
	cfg := Config{
		Transforms: []config.Config{
			{
				Type: "direct_assignment",
				Settings: map[string]interface{}{
					"source": "$.message",
					"target": "$.foo",
				},
			},
		},
	}

	// Create vibestation instance
	vibe, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create vibestation: %v", err)
	}

	// Create test message
	msg := message.New()
	msg.SetData([]byte(`{"message": "hello world", "value": 42}`))

	// Transform the message
	result, err := vibe.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result message, got %d", len(result))
	}

	// Check that the field was copied
	data := string(result[0].Data())
	expected := `{"message":"hello world","value":42,"foo":"hello world"}`
	if !jsonEqual(data, expected) {
		t.Errorf("Expected %s, got %s", expected, data)
	}
}
