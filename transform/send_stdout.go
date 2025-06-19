package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

type sendStdoutConfig struct {
	ID string `json:"id"`
}

func (c *sendStdoutConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func newSendStdout(_ context.Context, cfg config.Config) (*sendStdout, error) {
	conf := sendStdoutConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform send_stdout: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "send_stdout"
	}

	tf := sendStdout{
		conf: conf,
	}

	return &tf, nil
}

type sendStdout struct {
	conf sendStdoutConfig
	mu   sync.Mutex
}

func (tf *sendStdout) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	// Print the message data to stdout
	fmt.Println(string(msg.Data()))

	return []*message.Message{msg}, nil
}

func (tf *sendStdout) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
