package transform

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

type DecompressGzipConfig struct {
	ID string `json:"id"`
}

func (c *DecompressGzipConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func newDecompressGzip(_ context.Context, cfg config.Config) (*DecompressGzip, error) {
	conf := DecompressGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform decompress_gzip: %v", err)
	}

	// Use settings to determine ID (named only)
	id := "decompress_gzip"
	if v, ok := cfg.Settings["id"]; ok {
		if s, ok := v.(string); ok && s != "" {
			id = s
		}
	}
	conf.ID = id

	// Universal source argument (named only)
	var sourcePath string
	if v, ok := cfg.Settings["source"]; ok {
		if s, ok := v.(string); ok {
			sourcePath = s
		}
	}

	// Target path for assignments
	var targetPath string
	if v, ok := cfg.Settings["target"]; ok {
		if s, ok := v.(string); ok {
			targetPath = s
		}
	}

	tf := DecompressGzip{
		conf:       conf,
		settings:   cfg.Settings,
		sourcePath: sourcePath,
		targetPath: targetPath,
	}

	return &tf, nil
}

type DecompressGzip struct {
	conf       DecompressGzipConfig
	settings   map[string]interface{}
	sourcePath string
	targetPath string
}

func (tf *DecompressGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// Determine input data
	var inputData []byte
	if tf.sourcePath != "" {
		val := msg.GetPathValue(tf.sourcePath)
		if val.Exists() {
			inputData = val.Bytes()
		}
	}
	if inputData == nil {
		inputData = msg.Data()
	}

	decompressed, err := decompressGzip(inputData)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If targetPath is set, store the result in the target JSON path
	if tf.targetPath != "" {
		err := msg.SetPathValue(tf.targetPath, string(decompressed))
		if err != nil {
			return nil, fmt.Errorf("transform %s: failed to set target: %v", tf.conf.ID, err)
		}
	} else {
		msg.SetData(decompressed)
	}

	return []*message.Message{msg}, nil
}

func (tf *DecompressGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

// decompressGzip decompresses gzipped data.
func decompressGzip(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}
