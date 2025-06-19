package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

type SendStdoutConfig struct {
	ID string `json:"id"`
}

func (c *SendStdoutConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func newSendStdout(_ context.Context, cfg config.Config) (*SendStdout, error) {
	conf := SendStdoutConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_stdout: %v", err)
	}

	// Use settings to determine ID (named only)
	id := "send_stdout"
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

	tf := SendStdout{
		conf:       conf,
		settings:   cfg.Settings,
		sourcePath: sourcePath,
		targetPath: targetPath,
	}

	return &tf, nil
}

type SendStdout struct {
	conf       SendStdoutConfig
	settings   map[string]interface{}
	sourcePath string
	targetPath string
	mu         sync.Mutex
}

func (tf *SendStdout) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// Determine input data
	var inputData []byte
	if tf.sourcePath != "" {
		val := msg.GetValue(tf.sourcePath)
		if val.Exists() {
			inputData = val.Bytes()
		}
	}
	if inputData == nil {
		inputData = msg.Data()
	}

	// If targetPath is set, store the input in the target JSON path
	if tf.targetPath != "" {
		err := msg.SetValue(tf.targetPath, string(inputData))
		if err != nil {
			return nil, fmt.Errorf("transform %s: failed to set target: %v", tf.conf.ID, err)
		}
	}

	// Print the message data to stdout
	fmt.Println(string(inputData))

	return []*message.Message{msg}, nil
}

func (tf *SendStdout) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
