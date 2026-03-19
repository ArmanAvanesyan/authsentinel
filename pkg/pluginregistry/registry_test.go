package pluginregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
)

func noopFactory(context.Context, pluginapi.PluginDescriptor) (pluginapi.Plugin, error) {
	return nil, nil
}

func TestRegister_Success(t *testing.T) {
	r := New()
	desc := pluginapi.PluginDescriptor{
		ID: "test-plugin", Kind: pluginapi.PluginKindIntegration,
		Name: "Test", Version: "0.0.0", Capabilities: []pluginapi.Capability{"integration:test"},
	}
	err := r.Register(desc, noopFactory)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	reg, ok := r.RegistrationFor("test-plugin")
	if !ok || reg == nil {
		t.Fatal("RegistrationFor: expected registration")
	}
	if reg.Descriptor.ID != desc.ID || !reg.Enabled {
		t.Errorf("descriptor id=%q enabled=%v", reg.Descriptor.ID, reg.Enabled)
	}
	ids := r.AllPluginIDs()
	if len(ids) != 1 || ids[0] != "test-plugin" {
		t.Errorf("AllPluginIDs: got %v", ids)
	}
}

func TestRegister_DuplicateID_ReturnsErrAlreadyRegistered(t *testing.T) {
	r := New()
	desc := pluginapi.PluginDescriptor{
		ID: "dup", Kind: pluginapi.PluginKindPipeline,
		Name: "Dup", Version: "0.0.0", Capabilities: []pluginapi.Capability{"pipeline:dup"},
	}
	if err := r.Register(desc, noopFactory); err != nil {
		t.Fatalf("first Register: %v", err)
	}
	err := r.Register(desc, noopFactory)
	if err == nil {
		t.Fatal("expected error on duplicate Register")
	}
	if !isErr(err, ErrAlreadyRegistered) {
		t.Errorf("expected ErrAlreadyRegistered, got %v", err)
	}
}

func TestEnable_Disable_UnknownPlugin_ReturnsErrUnknownPlugin(t *testing.T) {
	r := New()
	if err := r.Enable("missing"); err == nil || !isErr(err, ErrUnknownPlugin) {
		t.Errorf("Enable(missing): want ErrUnknownPlugin, got %v", err)
	}
	if err := r.Disable("missing"); err == nil || !isErr(err, ErrUnknownPlugin) {
		t.Errorf("Disable(missing): want ErrUnknownPlugin, got %v", err)
	}
}

func TestEnable_Disable_ResolveByCapability(t *testing.T) {
	r := New()
	cap := pluginapi.Capability("pipeline:foo")
	desc := pluginapi.PluginDescriptor{
		ID: "foo", Kind: pluginapi.PluginKindPipeline,
		Name: "Foo", Version: "1.0", Capabilities: []pluginapi.Capability{cap},
	}
	if err := r.Register(desc, noopFactory); err != nil {
		t.Fatalf("Register: %v", err)
	}
	got := r.ResolveByCapability(cap)
	if len(got) != 1 || got[0].ID != "foo" {
		t.Errorf("ResolveByCapability: got %v", got)
	}
	if err := r.Disable("foo"); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	got = r.ResolveByCapability(cap)
	if len(got) != 0 {
		t.Errorf("after Disable, ResolveByCapability: got %v", got)
	}
	if err := r.Enable("foo"); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	got = r.ResolveByCapability(cap)
	if len(got) != 1 || got[0].ID != "foo" {
		t.Errorf("after Enable, ResolveByCapability: got %v", got)
	}
}

func TestResolveAllByKind(t *testing.T) {
	r := New()
	for _, d := range []pluginapi.PluginDescriptor{
		{ID: "p1", Kind: pluginapi.PluginKindPipeline, Name: "P1", Version: "0", Capabilities: []pluginapi.Capability{"pipeline:a"}},
		{ID: "p2", Kind: pluginapi.PluginKindPipeline, Name: "P2", Version: "0", Capabilities: []pluginapi.Capability{"pipeline:b"}},
		{ID: "i1", Kind: pluginapi.PluginKindIntegration, Name: "I1", Version: "0", Capabilities: []pluginapi.Capability{"integration:x"}},
	} {
		if err := r.Register(d, noopFactory); err != nil {
			t.Fatalf("Register %s: %v", d.ID, err)
		}
	}
	pipeline := r.ResolveAllByKind(pluginapi.PluginKindPipeline)
	if len(pipeline) != 2 {
		t.Fatalf("ResolveAllByKind(pipeline): got %d, want 2", len(pipeline))
	}
	integration := r.ResolveAllByKind(pluginapi.PluginKindIntegration)
	if len(integration) != 1 || integration[0].ID != "i1" {
		t.Errorf("ResolveAllByKind(integration): got %v", integration)
	}
	if err := r.Disable("p1"); err != nil {
		t.Fatalf("Disable(p1): %v", err)
	}
	pipeline = r.ResolveAllByKind(pluginapi.PluginKindPipeline)
	if len(pipeline) != 1 || pipeline[0].ID != "p2" {
		t.Errorf("after Disable(p1), ResolveAllByKind(pipeline): got %v", pipeline)
	}
}

func TestBuildDependencyGraph_Success(t *testing.T) {
	capProvider := pluginapi.Capability("provider:oidc")
	capPipeline := pluginapi.Capability("pipeline:auth")
	r := New()
	// Provider first (no deps), then pipeline depends on provider.
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "oidc", Kind: pluginapi.PluginKindProvider,
		Name: "OIDC", Version: "0", Capabilities: []pluginapi.Capability{capProvider},
	}, noopFactory); err != nil {
		t.Fatalf("Register oidc: %v", err)
	}
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "auth-pipeline", Kind: pluginapi.PluginKindPipeline,
		Name: "Auth", Version: "0", Capabilities: []pluginapi.Capability{capPipeline},
		DependsOn: []pluginapi.Capability{capProvider},
	}, noopFactory); err != nil {
		t.Fatalf("Register auth-pipeline: %v", err)
	}
	if err := r.BuildDependencyGraph(); err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}
	order := r.StartupOrder()
	if len(order) != 2 {
		t.Fatalf("StartupOrder: got %v", order)
	}
	// Provider (no deps) must come before auth-pipeline.
	if order[0] != "oidc" || order[1] != "auth-pipeline" {
		t.Errorf("StartupOrder: want [oidc auth-pipeline], got %v", order)
	}
}

func TestBuildDependencyGraph_MissingCapability(t *testing.T) {
	r := New()
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "needs-oidc", Kind: pluginapi.PluginKindPipeline,
		Name: "NeedsOIDC", Version: "0", Capabilities: []pluginapi.Capability{"pipeline:x"},
		DependsOn: []pluginapi.Capability{"provider:oidc"},
	}, noopFactory); err != nil {
		t.Fatalf("Register: %v", err)
	}
	err := r.BuildDependencyGraph()
	if err == nil {
		t.Fatal("expected error when dependency capability has no providers")
	}
}

func TestBuildDependencyGraph_Cycle_ReturnsErrDependencyCycle(t *testing.T) {
	capA := pluginapi.Capability("cap:a")
	capB := pluginapi.Capability("cap:b")
	r := New()
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "plugin-a", Kind: pluginapi.PluginKindPipeline,
		Name: "A", Version: "0", Capabilities: []pluginapi.Capability{capA},
		DependsOn: []pluginapi.Capability{capB},
	}, noopFactory); err != nil {
		t.Fatalf("Register A: %v", err)
	}
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "plugin-b", Kind: pluginapi.PluginKindPipeline,
		Name: "B", Version: "0", Capabilities: []pluginapi.Capability{capB},
		DependsOn: []pluginapi.Capability{capA},
	}, noopFactory); err != nil {
		t.Fatalf("Register B: %v", err)
	}
	err := r.BuildDependencyGraph()
	if err == nil {
		t.Fatal("expected error for dependency cycle")
	}
	if !isErr(err, ErrDependencyCycle) {
		t.Errorf("expected ErrDependencyCycle, got %v", err)
	}
}

func TestStartupOrder_EmptyBeforeBuildDependencyGraph(t *testing.T) {
	r := New()
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "x", Kind: pluginapi.PluginKindPipeline,
		Name: "X", Version: "0", Capabilities: []pluginapi.Capability{"pipeline:x"},
	}, noopFactory); err != nil {
		t.Fatalf("Register: %v", err)
	}
	order := r.StartupOrder()
	if len(order) != 0 {
		t.Errorf("StartupOrder before BuildDependencyGraph: got %v", order)
	}
	if err := r.BuildDependencyGraph(); err != nil {
		t.Fatalf("BuildDependencyGraph: %v", err)
	}
	order = r.StartupOrder()
	if len(order) != 1 || order[0] != "x" {
		t.Errorf("StartupOrder after Build: got %v", order)
	}
}

func TestRegistrationFor(t *testing.T) {
	r := New()
	if _, ok := r.RegistrationFor("none"); ok {
		t.Fatal("RegistrationFor(none): expected false")
	}
	if err := r.Register(pluginapi.PluginDescriptor{
		ID: "r1", Kind: pluginapi.PluginKindProvider,
		Name: "R1", Version: "0", Capabilities: []pluginapi.Capability{"provider:r1"},
	}, noopFactory); err != nil {
		t.Fatalf("Register: %v", err)
	}
	reg, ok := r.RegistrationFor("r1")
	if !ok || reg == nil || reg.Descriptor.ID != "r1" {
		t.Errorf("RegistrationFor(r1): got ok=%v reg=%v", ok, reg)
	}
}

func isErr(err, target error) bool {
	for err != nil {
		if errors.Is(err, target) {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}
