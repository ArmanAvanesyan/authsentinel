package policy

// Decision is the result of evaluating a policy against an input.
type Decision struct {
	Allow      bool
	StatusCode int
	Headers    map[string]string
	Reason     string
}
