package transform

import (
	"context"
	"testing"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

func TestLowercaseStringTransform_Basic(t *testing.T) {
	cfg := config.Config{
		Type: "lowercase_string",
		Settings: map[string]interface{}{
			"source": "$.foo",
		},
	}

	tf, err := newLowercaseString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create lowercase_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"foo": "HELLO WORLD"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	expected := "hello world"
	if string(msgs[0].Data()) != expected {
		t.Errorf("expected %q, got %q", expected, string(msgs[0].Data()))
	}
}

func TestLowercaseStringTransform_WithTarget(t *testing.T) {
	cfg := config.Config{
		Type: "lowercase_string",
		Settings: map[string]interface{}{
			"source": "$.foo",
			"target": "$.bar",
		},
	}

	tf, err := newLowercaseString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create lowercase_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"foo": "HELLO WORLD"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	val := msgs[0].GetPathValue("bar")
	if !val.Exists() {
		t.Fatal("expected bar value to exist in target path")
	}
	if val.String() != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", val.String())
	}
}

func TestLowercaseStringTransform_NoSource(t *testing.T) {
	cfg := config.Config{
		Type:     "lowercase_string",
		Settings: map[string]interface{}{},
	}

	tf, err := newLowercaseString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create lowercase_string transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte("HELLO WORLD"))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	expected := "hello world"
	if string(msgs[0].Data()) != expected {
		t.Errorf("expected %q, got %q", expected, string(msgs[0].Data()))
	}
}

func TestLowercaseStringTransform_ControlMessage(t *testing.T) {
	cfg := config.Config{
		Type: "lowercase_string",
		Settings: map[string]interface{}{
			"source": "$.foo",
		},
	}

	tf, err := newLowercaseString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create lowercase_string transform: %v", err)
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
