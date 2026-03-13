package policy

// RegoEvaluator runs policy bundles written in Rego (e.g. OPA-style).
// TODO: implement Rego evaluation (embed OPA/Rego engine or call out to one).
type RegoEvaluator struct{}

// Load loads a Rego policy bundle from path.
func (r *RegoEvaluator) Load(path string) error {
	return nil
}
