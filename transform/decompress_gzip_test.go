package transform

import (
	"bytes"
	"compress/gzip"
	"context"
	"testing"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

func TestDecompressGzipTransform_Basic(t *testing.T) {
	// Create gzipped test data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("test data"))
	gw.Close()
	gzippedData := buf.Bytes()

	cfg := config.Config{
		Type: "decompress_gzip",
		Settings: map[string]interface{}{
			"source": "$.compressed",
		},
	}

	tf, err := newDecompressGzip(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decompress_gzip transform: %v", err)
	}

	msg := message.New()
	msg.SetData(gzippedData)

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

func TestDecompressGzipTransform_WithTarget(t *testing.T) {
	// Create gzipped test data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("test data"))
	gw.Close()
	gzippedData := buf.Bytes()

	cfg := config.Config{
		Type: "decompress_gzip",
		Settings: map[string]interface{}{
			"source": "$.compressed",
			"target": "$.decompressed",
		},
	}

	tf, err := newDecompressGzip(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decompress_gzip transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"compressed": "` + string(gzippedData) + `"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error when using target with non-JSON data, got nil")
	}
	if msgs != nil {
		t.Errorf("expected no messages on error, got %v", msgs)
	}
}

func TestDecompressGzipTransform_NoSource(t *testing.T) {
	// Create gzipped test data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("test data"))
	gw.Close()
	gzippedData := buf.Bytes()

	cfg := config.Config{
		Type:     "decompress_gzip",
		Settings: map[string]interface{}{},
	}

	tf, err := newDecompressGzip(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decompress_gzip transform: %v", err)
	}

	msg := message.New()
	msg.SetData(gzippedData)

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

func TestDecompressGzipTransform_InvalidGzip(t *testing.T) {
	cfg := config.Config{
		Type: "decompress_gzip",
		Settings: map[string]interface{}{
			"source": "$.compressed",
		},
	}

	tf, err := newDecompressGzip(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decompress_gzip transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"compressed": "not_gzipped_data"}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error for invalid gzip data, got nil")
	}
	if msgs != nil {
		t.Errorf("expected no messages on error, got %v", msgs)
	}
}

func TestDecompressGzipTransform_EmptyData(t *testing.T) {
	cfg := config.Config{
		Type: "decompress_gzip",
		Settings: map[string]interface{}{
			"source": "$.compressed",
		},
	}

	tf, err := newDecompressGzip(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decompress_gzip transform: %v", err)
	}

	msg := message.New()
	msg.SetData([]byte(`{"compressed": ""}`))

	msgs, err := tf.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	if len(msgs[0].Data()) != 0 {
		t.Errorf("expected empty data, got %q", string(msgs[0].Data()))
	}
}

func TestDecompressGzipTransform_ControlMessage(t *testing.T) {
	cfg := config.Config{
		Type: "decompress_gzip",
		Settings: map[string]interface{}{
			"source": "$.compressed",
		},
	}

	tf, err := newDecompressGzip(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create decompress_gzip transform: %v", err)
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
