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

type formatGzipConfig struct {
	ID string `json:"id"`
}

func (c *formatGzipConfig) Decode(in interface{}) error {
	if in == nil {
		return nil
	}

	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, c)
}

func newFormatFromGzip(_ context.Context, cfg config.Config) (*formatFromGzip, error) {
	conf := formatGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform format_from_gzip: %v", err)
	}

	if conf.ID == "" {
		conf.ID = "format_from_gzip"
	}

	tf := formatFromGzip{
		conf: conf,
	}

	return &tf, nil
}

type formatFromGzip struct {
	conf formatGzipConfig
}

func (tf *formatFromGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	decompressed, err := fmtFromGzip(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform %s: %v", tf.conf.ID, err)
	}

	msg.SetData(decompressed)
	return []*message.Message{msg}, nil
}

func (tf *formatFromGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}

// fmtFromGzip decompresses gzipped data.
func fmtFromGzip(data []byte) ([]byte, error) {
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
