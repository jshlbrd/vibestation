package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

type stringSplitConfig struct {
	// Separator splits the string into elements of the array.
	Separator string `json:"separator"`

	ID string `json:"id"`
}

func (c *stringSplitConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func (c *stringSplitConfig) Validate() error {
	if c.Separator == "" {
		return fmt.Errorf("separator: missing required option")
	}

	return nil
}

func newStringSplit(_ context.Context, cfg config.Config) (*stringSplit, error) {
	conf := stringSplitConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform string_split: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "string_split"
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("transform %s: %v", conf.ID, err)
	}

	tf := stringSplit{
		conf:      conf,
		separator: []byte(conf.Separator),
	}

	return &tf, nil
}

type stringSplit struct {
	conf      stringSplitConfig
	separator []byte
}

func (tf *stringSplit) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// Split the data by the separator
	parts := bytes.Split(msg.Data(), tf.separator)

	// Create a new message for each part
	var result []*message.Message
	for _, part := range parts {
		// Skip empty parts
		if len(part) == 0 {
			continue
		}

		newMsg := message.New().SetData(part)
		result = append(result, newMsg)
	}

	return result, nil
}

func (tf *stringSplit) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
