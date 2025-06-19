// Package transform provides functions for transforming messages.
package transform

import (
	"context"
	"fmt"

	"github.com/josh.liburdi/vibestation/config"
	"github.com/josh.liburdi/vibestation/message"
)

// Transformer is the interface implemented by all transforms and
// provides the ability to transform a message.
type Transformer interface {
	Transform(context.Context, *message.Message) ([]*message.Message, error)
}

// Factory can be used to implement custom transform factory functions.
type Factory func(context.Context, config.Config) (Transformer, error)

// New is a factory function for returning a configured Transformer.
func New(ctx context.Context, cfg config.Config) (Transformer, error) {
	switch cfg.Type {
	case "decompress_gzip":
		return newDecompressGzip(ctx, cfg)
	case "split_string":
		return newSplitString(ctx, cfg)
	case "send_stdout":
		return newSendStdout(ctx, cfg)
	case "decode_base64":
		return newDecodeBase64(ctx, cfg)
	case "lowercase_string":
		return newLowercaseString(ctx, cfg)
	default:
		return nil, fmt.Errorf("transform %s: unsupported transform type", cfg.Type)
	}
}

// Apply applies one or more transform functions to one or more messages.
func Apply(ctx context.Context, tf []Transformer, msgs ...*message.Message) ([]*message.Message, error) {
	resultMsgs := make([]*message.Message, len(msgs))
	copy(resultMsgs, msgs)

	for i := 0; len(resultMsgs) > 0 && i < len(tf); i++ {
		var nextResultMsgs []*message.Message
		for _, m := range resultMsgs {
			rMsgs, err := tf[i].Transform(ctx, m)
			if err != nil {
				// We immediately return if a transform hits an unrecoverable
				// error on a message.
				return nil, err
			}
			nextResultMsgs = append(nextResultMsgs, rMsgs...)
		}
		resultMsgs = nextResultMsgs
	}

	return resultMsgs, nil
}
