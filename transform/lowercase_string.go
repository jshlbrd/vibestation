package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

type LowercaseStringConfig struct {
	ID string `json:"id"`
}

func (c *LowercaseStringConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, c)
}

func newLowercaseString(_ context.Context, cfg config.Config) (*LowercaseStringTransform, error) {
	conf := LowercaseStringConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform lowercase_string: %v", err)
	}

	id := "lowercase_string"
	if v, ok := cfg.Settings["id"]; ok {
		if s, ok := v.(string); ok && s != "" {
			id = s
		}
	}
	conf.ID = id

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

	tf := LowercaseStringTransform{
		conf:       conf,
		sourcePath: sourcePath,
		targetPath: targetPath,
		settings:   cfg.Settings,
	}

	return &tf, nil
}

type LowercaseStringTransform struct {
	conf       LowercaseStringConfig
	sourcePath string
	targetPath string
	settings   map[string]interface{}
}

func (tf *LowercaseStringTransform) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	var input string
	if tf.sourcePath != "" {
		val := msg.GetPathValue(tf.sourcePath)
		if val.Exists() {
			input = val.String()
		}
	}
	if input == "" {
		input = string(msg.Data())
	}

	lower := strings.ToLower(input)

	if tf.targetPath != "" {
		err := msg.SetPathValue(tf.targetPath, lower)
		if err != nil {
			return nil, fmt.Errorf("transform %s: failed to set target: %v", tf.conf.ID, err)
		}
	} else {
		msg.SetData([]byte(lower))
	}

	return []*message.Message{msg}, nil
}

func (tf *LowercaseStringTransform) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
