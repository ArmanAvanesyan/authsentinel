package httpserver

import "testing"

func TestNewReturnsServerWithHandler(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatalf("expected non-nil Server")
	}
	if s.Handler() == nil {
		t.Fatalf("expected non-nil Handler")
	}
}
