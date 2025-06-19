package transform

import (
	"context"
	"reflect"
	"testing"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

func TestSplitString_Basic(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "\n",
		},
	}
	ts, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}
	msg := message.New().SetData([]byte("a\nb\nc"))
	results, err := ts.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	expected := []string{"a", "b", "c"}
	for i, r := range results {
		if string(r.Data()) != expected[i] {
			t.Errorf("expected '%s', got '%s'", expected[i], string(r.Data()))
		}
	}
}

func TestSplitString_SourceTarget(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": ",",
			"source":    "$.foo",
			"target":    "$.bar",
		},
	}
	ts, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}
	msg := message.New().SetData([]byte(`{"foo": "x,y,z"}`))
	results, err := ts.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	expected := []string{"x", "y", "z"}
	for i, r := range results {
		val := r.GetValue("$.bar")
		if !val.Exists() || val.String() != expected[i] {
			t.Errorf("expected bar='%s', got '%s'", expected[i], val.String())
		}
	}
}

func TestSplitString_EmptyInput(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "|",
		},
	}
	ts, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}
	msg := message.New().SetData([]byte(""))
	results, err := ts.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSplitString_ControlMessage(t *testing.T) {
	cfg := config.Config{
		Type: "split_string",
		Settings: map[string]interface{}{
			"separator": "|",
		},
	}
	ts, err := newSplitString(context.Background(), cfg)
	if err != nil {
		t.Fatalf("failed to create split_string transform: %v", err)
	}
	msg := message.New().AsControl()
	results, err := ts.Transform(context.Background(), msg)
	if err != nil {
		t.Fatalf("transform failed: %v", err)
	}
	if len(results) != 1 || !reflect.DeepEqual(results[0], msg) {
		t.Errorf("expected control message to be passed through unchanged")
	}
}
