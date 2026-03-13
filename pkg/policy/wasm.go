package policy

// WASMRuntime runs policy bundles compiled to WASM.
// TODO: implement embedded WASM runtime.
type WASMRuntime struct{}

// Load loads a policy bundle from path.
func (w *WASMRuntime) Load(path string) error {
	return nil
}
