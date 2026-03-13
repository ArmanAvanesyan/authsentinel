package policy

import "context"

// Engine evaluates authorization decisions using embedded policy bundles.
type Engine interface {
	Evaluate(ctx context.Context, input Input) (*Decision, error)
}
