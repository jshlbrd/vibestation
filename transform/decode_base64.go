package transform

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

type DecodeBase64Config struct {
	ID string `json:"id"`
}

func (c *DecodeBase64Config) Decode(in interface{}) error {
	if in == nil {
		return nil
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func newDecodeBase64(_ context.Context, cfg config.Config) (*DecodeBase64Transform, error) {
	conf := DecodeBase64Config{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform decode_base64: %v", err)
	}

	// Use settings to determine ID (named only)
	id := "decode_base64"
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

	tf := DecodeBase64Transform{
		conf:       conf,
		settings:   cfg.Settings,
		sourcePath: sourcePath,
		targetPath: targetPath,
	}

	return &tf, nil
}

type DecodeBase64Transform struct {
	conf       DecodeBase64Config
	settings   map[string]interface{}
	sourcePath string
	targetPath string
}

func (tf *DecodeBase64Transform) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
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

	decoded, err := decodeBase64(inputData)
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	// If we have a target path, store the result there
	if tf.targetPath != "" {
		err := msg.SetPathValue(tf.targetPath, string(decoded))
		if err != nil {
			return nil, fmt.Errorf("transform %s: failed to set target: %v", tf.conf.ID, err)
		}
	} else {
		// Otherwise, set as message data
		msg.SetData(decoded)
	}

	return []*message.Message{msg}, nil
}

func (tf *DecodeBase64Transform) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

// decodeBase64 decodes base64-encoded data.
func decodeBase64(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	// Convert to string and trim whitespace
	input := strings.TrimSpace(string(data))

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %v", err)
	}

	return decoded, nil
}
