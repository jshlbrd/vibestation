package transform

import (
	"context"
	"testing"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

func TestDecodeBase64Transform_InvalidBase64(t *testing.T) {
	cfg := config.Config{
		Type: "decode_base64",
		Settings: map[string]interface{}{
			"source": "$.foo",
		},
	}

	tf, err := newDecodeBase64(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decode_base64 transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"foo": "not_base64!"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error for invalid base64, got nil")
	}
	if msgs != nil {
		t.Errorf("expected no messages on error, got %v", msgs)
	}
}

func TestDecodeBase64Transform_ValidBase64(t *testing.T) {
	cfg := config.Config{
		Type: "decode_base64",
		Settings: map[string]interface{}{
			"source": "$.encoded",
		},
	}

	tf, err := newDecodeBase64(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decode_base64 transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"encoded": "dGVzdCBkYXRh"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	expected := "test data"
	if string(msgs[0].Data()) != expected {
		t.Errorf("expected %q, got %q", expected, string(msgs[0].Data()))
	}
}

func TestDecodeBase64Transform_WithTarget(t *testing.T) {
	cfg := config.Config{
		Type: "decode_base64",
		Settings: map[string]interface{}{
			"source": "$.encoded",
			"target": "$.decoded",
		},
	}

	tf, err := newDecodeBase64(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decode_base64 transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"encoded": "dGVzdCBkYXRh"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	val := msgs[0].GetValue("decoded")
	if !val.Exists() {
		t.Fatal("expected decoded value to exist in target path")
	}
	if val.String() != "test data" {
		t.Errorf("expected %q, got %q", "test data", val.String())
	}
}

func TestDecodeBase64Transform_NoSource(t *testing.T) {
	cfg := config.Config{
		Type:     "decode_base64",
		Settings: map[string]interface{}{},
	}

	tf, err := newDecodeBase64(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decode_base64 transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte("dGVzdCBkYXRh"))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	expected := "test data"
	if string(msgs[0].Data()) != expected {
		t.Errorf("expected %q, got %q", expected, string(msgs[0].Data()))
	}
}

func TestDecodeBase64Transform_ControlMessage(t *testing.T) {
	cfg := config.Config{
		Type: "decode_base64",
		Settings: map[string]interface{}{
			"source": "$.encoded",
		},
	}

	tf, err := newDecodeBase64(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decode_base64 transform: %v", err)
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
