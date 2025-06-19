package transform

import (
	"context"
	"testing"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

func TestSplitStringTransform_Basic(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "\n",
			"source":    "$.data",
		},
	}

	tf, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "line1\nline2\nline3"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	expected := []string{"line1", "line2", "line3"}
	for i, expectedLine := range expected {
		if string(msgs[i].Data()) != expectedLine {
			t.Errorf("message %d: expected %q, got %q", i, expectedLine, string(msgs[i].Data()))
		}
	}
}

func TestSplitStringTransform_WithTarget(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "\n",
			"source":    "$.data",
			"target":    "$.lines",
		},
	}

	tf, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "line1\nline2"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	// Check that each message has the line stored in the target path
	for i, expectedLine := range []string{"line1", "line2"} {
		val := msgs[i].GetValue("$.lines")
		if !val.Exists() {
			t.Errorf("message %d: expected line to exist in target path", i)
		}
		if val.String() != expectedLine {
			t.Errorf("message %d: expected %q, got %q", i, expectedLine, val.String())
		}
	}
}

func TestSplitStringTransform_CustomSeparator(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "|",
			"source":    "$.data",
		},
	}

	tf, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "part1|part2|part3"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	expected := []string{"part1", "part2", "part3"}
	for i, expectedPart := range expected {
		if string(msgs[i].Data()) != expectedPart {
			t.Errorf("message %d: expected %q, got %q", i, expectedPart, string(msgs[i].Data()))
		}
	}
}

func TestSplitStringTransform_NoSource(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "\n",
		},
	}

	tf, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte("line1\nline2\nline3"))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	expected := []string{"line1", "line2", "line3"}
	for i, expectedLine := range expected {
		if string(msgs[i].Data()) != expectedLine {
			t.Errorf("message %d: expected %q, got %q", i, expectedLine, string(msgs[i].Data()))
		}
	}
}

func TestSplitStringTransform_EmptyLines(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "\n",
			"source":    "$.data",
		},
	}

	tf, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "line1\n\nline3"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages (empty lines should be skipped), got %d", len(msgs))
	}

	expected := []string{"line1", "line3"}
	for i, expectedLine := range expected {
		if string(msgs[i].Data()) != expectedLine {
			t.Errorf("message %d: expected %q, got %q", i, expectedLine, string(msgs[i].Data()))
		}
	}
}

func TestSplitStringTransform_ControlMessage(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "\n",
			"source":    "$.data",
		},
	}

	tf, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}

	msg := message.New().AsControl()

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if !msgs[0].IsControl() {
		t.Error("expected control message to remain control message")
	}
}

func TestSplitStringTransform_ValidationError(t *testing.T) {
	cfg := config.Config{
		Type:     "split_string",
		Settings: map[string]interface{}{
			// Missing separator should cause validation error
		},
	}

	_, err := newSplitString(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected validation error for missing separator, got nil")
	}
}
