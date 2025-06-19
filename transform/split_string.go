package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
)

type SplitStringConfig struct {
	Separator string `json:"separator"`
	ID        string `json:"id"`
}

func (c *SplitStringConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, c)
}

func (c *SplitStringConfig) Validate() error {
	if c.Separator == "" {
		return fmt.Errorf("separator: missing required option")
	}
	return nil
}

func newSplitString(_ context.Context, cfg config.Config) (*SplitString, error) {
	conf := SplitStringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform split_string: %v", err)
	}
	if conf.ID == "" {
		conf.ID = "split_string"
	}
	separator := "\n"
	if sep, ok := cfg.Settings["separator"]; ok {
		if s, ok := sep.(string); ok {
			separator = s
		}
	}
	var sourcePath string
	if v, ok := cfg.Settings["source"]; ok {
		if s, ok := v.(string); ok {
			sourcePath = s
		}
	}
	var targetPath string
	if v, ok := cfg.Settings["target"]; ok {
		if s, ok := v.(string); ok {
			targetPath = s
		}
	}
	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}
	tf := SplitString{
		conf:       conf,
		separator:  []byte(separator),
		settings:   cfg.Settings,
		sourcePath: sourcePath,
		targetPath: targetPath,
	}
	return &tf, nil
}

type SplitString struct {
	conf       SplitStringConfig
	separator  []byte
	settings   map[string]interface{}
	sourcePath string
	targetPath string
}

func (tf *SplitString) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}
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
	parts := bytes.Split(inputData, tf.separator)
	var result []*message.Message
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		var newMsg *message.Message
		if tf.targetPath != "" {
			newMsg = message.New().SetData([]byte("{}"))
			err := newMsg.SetValue(tf.targetPath, string(part))
			if err != nil {
				return nil, fmt.Errorf("transform %s: failed to set target: %v", tf.conf.ID, err)
			}
		} else {
			newMsg = message.New().SetData(part)
		}
		result = append(result, newMsg)
	}
	return result, nil
}

func (tf *SplitString) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
