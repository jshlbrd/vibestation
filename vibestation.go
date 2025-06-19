package vibestation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jshlbrd/vibestation/config"
	"github.com/jshlbrd/vibestation/message"
	"github.com/jshlbrd/vibestation/transform"
)

var errNoTransforms = fmt.Errorf("no transforms configured")

// Config is the core configuration for the application. Custom applications
// should embed this and add additional configuration options.
type Config struct {
	// Transforms contains a list of data transformations that are executed.
	Transforms []config.Config `json:"transforms"`
}

// Vibestation provides access to data transformation functions.
type Vibestation struct {
	cfg Config

	factory transform.Factory
	tforms  []transform.Transformer
}

// New returns a new Vibestation instance.
func New(ctx context.Context, cfg Config, opts ...func(*Vibestation)) (*Vibestation, error) {
	if cfg.Transforms == nil {
		return nil, errNoTransforms
	}

	vibe := &Vibestation{
		cfg:     cfg,
		factory: transform.New,
	}

	for _, o := range opts {
		o(vibe)
	}

	// Create transforms from the configuration.
	for _, c := range cfg.Transforms {
		t, err := vibe.factory(ctx, c)
		if err != nil {
			return nil, err
		}

		vibe.tforms = append(vibe.tforms, t)
	}

	return vibe, nil
}

// WithTransformFactory implements a custom transform factory.
func WithTransformFactory(fac transform.Factory) func(*Vibestation) {
	return func(v *Vibestation) {
		v.factory = fac
	}
}

// Transform runs the configured data transformation functions on the
// provided messages.
//
// This is safe to use concurrently.
func (v *Vibestation) Transform(ctx context.Context, msg ...*message.Message) ([]*message.Message, error) {
	return transform.Apply(ctx, v.tforms, msg...)
}

// String returns a JSON representation of the configuration.
func (v *Vibestation) String() string {
	b, err := json.Marshal(v.cfg)
	if err != nil {
		return fmt.Sprintf("vibestation: %v", err)
	}

	return string(b)
}
