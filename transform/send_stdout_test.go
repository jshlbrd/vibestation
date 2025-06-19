package transform

import (
	"context"
	"testing"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

func TestSendStdoutTransform_Basic(t *testing.T) {
	cfg := config.Config{
		Type: "send_stdout",
		Settings: map[string]interface{}{
			"source": "$.data",
		},
	}

	tf, err := newSendStdout(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create send_stdout transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "test output"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// The message should be returned unchanged
	if string(msgs[0].Data()) != `{"data": "test output"}` {
		t.Errorf("expected message data to be unchanged, got %q", string(msgs[0].Data()))
	}
}

func TestSendStdoutTransform_WithTarget(t *testing.T) {
	cfg := config.Config{
		Type: "send_stdout",
		Settings: map[string]interface{}{
			"source": "$.data",
			"target": "$.output",
		},
	}

	tf, err := newSendStdout(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create send_stdout transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "test output"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// Check that the output value is stored in the target path
	val := msgs[0].GetValue("output")
	if !val.Exists() {
		t.Fatal("expected output value to exist in target path")
	}
	if val.String() != "test output" {
		t.Errorf("expected %q, got %q", "test output", val.String())
	}
}

func TestSendStdoutTransform_NoSource(t *testing.T) {
	cfg := config.Config{
		Type:     "send_stdout",
		Settings: map[string]interface{}{},
	}

	tf, err := newSendStdout(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create send_stdout transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte("test output"))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// The message should be returned unchanged
	if string(msgs[0].Data()) != "test output" {
		t.Errorf("expected message data to be unchanged, got %q", string(msgs[0].Data()))
	}
}

func TestSendStdoutTransform_EmptyData(t *testing.T) {
	cfg := config.Config{
		Type: "send_stdout",
		Settings: map[string]interface{}{
			"source": "$.data",
		},
	}

	tf, err := newSendStdout(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create send_stdout transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": ""}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// Empty data should be handled gracefully
	if string(msgs[0].Data()) != `{"data": ""}` {
		t.Errorf("expected message data to be unchanged, got %q", string(msgs[0].Data()))
	}
}

func TestSendStdoutTransform_ControlMessage(t *testing.T) {
	cfg := config.Config{
		Type: "send_stdout",
		Settings: map[string]interface{}{
			"source": "$.data",
		},
	}

	tf, err := newSendStdout(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create send_stdout transform: %v", err)
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

func TestSendStdoutTransform_NonExistentSource(t *testing.T) {
	cfg := config.Config{
		Type: "send_stdout",
		Settings: map[string]interface{}{
			"source": "$.nonexistent",
		},
	}

	tf, err := newSendStdout(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create send_stdout transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"data": "test"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// Should fall back to message data when source doesn't exist
	if string(msgs[0].Data()) != `{"data": "test"}` {
		t.Errorf("expected message data to be unchanged, got %q", string(msgs[0].Data()))
	}
}
