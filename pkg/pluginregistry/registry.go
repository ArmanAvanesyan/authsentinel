package pluginregistry

import (
	"context"
	"errors"
	"fmt"

	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
)

// Factory constructs a plugin instance for the given descriptor.
type Factory func(ctx context.Context, descriptor pluginapi.PluginDescriptor) (pluginapi.Plugin, error)

// Registration represents a registered plugin implementation and its runtime state.
type Registration struct {
	Descriptor pluginapi.PluginDescriptor
	Factory    Factory

	Enabled bool

	State pluginapi.PluginState
	Error error
}

// Registry holds registered plugins and provides resolution and lifecycle helpers.
type Registry struct {
	byID          map[pluginapi.PluginID]*Registration
	byCapability  map[pluginapi.Capability][]*Registration
	dependencies  map[pluginapi.PluginID][]pluginapi.Capability
	dependents    map[pluginapi.PluginID][]pluginapi.PluginID
	sortedStartup []pluginapi.PluginID
}

// New creates an empty Registry.
func New() *Registry {
	return &Registry{
		byID:         make(map[pluginapi.PluginID]*Registration),
		byCapability: make(map[pluginapi.Capability][]*Registration),
		dependencies: make(map[pluginapi.PluginID][]pluginapi.Capability),
		dependents:   make(map[pluginapi.PluginID][]pluginapi.PluginID),
	}
}

var (
	// ErrAlreadyRegistered is returned when attempting to register a plugin with a duplicate ID.
	ErrAlreadyRegistered = errors.New("pluginregistry: already registered")
	// ErrUnknownPlugin is returned when a referenced plugin is not known to the registry.
	ErrUnknownPlugin = errors.New("pluginregistry: unknown plugin")
	// ErrDependencyCycle is returned when plugin dependencies contain a cycle.
	ErrDependencyCycle = errors.New("pluginregistry: dependency cycle")
)

// Register adds a plugin descriptor and factory to the registry.
// Call BuildDependencyGraph after all registrations are complete.
func (r *Registry) Register(descriptor pluginapi.PluginDescriptor, factory Factory) error {
	if _, ok := r.byID[descriptor.ID]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, descriptor.ID)
	}
	reg := &Registration{
		Descriptor: descriptor,
		Factory:    factory,
		Enabled:    true,
		State:      pluginapi.PluginStateRegistered,
	}
	r.byID[descriptor.ID] = reg
	for _, cap := range descriptor.Capabilities {
		r.byCapability[cap] = append(r.byCapability[cap], reg)
	}
	if len(descriptor.DependsOn) > 0 {
		r.dependencies[descriptor.ID] = append([]pluginapi.Capability(nil), descriptor.DependsOn...)
	}
	return nil
}

// Enable marks the plugin with the given ID as enabled.
func (r *Registry) Enable(id pluginapi.PluginID) error {
	reg, ok := r.byID[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownPlugin, id)
	}
	reg.Enabled = true
	return nil
}

// Disable marks the plugin with the given ID as disabled.
func (r *Registry) Disable(id pluginapi.PluginID) error {
	reg, ok := r.byID[id]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownPlugin, id)
	}
	reg.Enabled = false
	return nil
}

// ResolveByCapability returns all enabled plugins that provide the given capability.
func (r *Registry) ResolveByCapability(cap pluginapi.Capability) []pluginapi.PluginDescriptor {
	regs := r.byCapability[cap]
	out := make([]pluginapi.PluginDescriptor, 0, len(regs))
	for _, reg := range regs {
		if reg.Enabled {
			out = append(out, reg.Descriptor)
		}
	}
	return out
}

// ResolveAllByKind returns all enabled plugins for the given kind.
func (r *Registry) ResolveAllByKind(kind pluginapi.PluginKind) []pluginapi.PluginDescriptor {
	out := make([]pluginapi.PluginDescriptor, 0, len(r.byID))
	for _, reg := range r.byID {
		if reg.Enabled && reg.Descriptor.Kind == kind {
			out = append(out, reg.Descriptor)
		}
	}
	return out
}

// BuildDependencyGraph validates plugin dependencies and computes an ordered startup list.
// Dependencies are expressed in terms of capabilities; this resolves them to concrete plugins.
func (r *Registry) BuildDependencyGraph() error {
	// Build dependents mapping by resolving capability dependencies to concrete plugin IDs.
	for id, caps := range r.dependencies {
		for _, cap := range caps {
			deps := r.byCapability[cap]
			if len(deps) == 0 {
				return fmt.Errorf("pluginregistry: plugin %s depends on capability %s with no providers", id, cap)
			}
			for _, reg := range deps {
				r.dependents[reg.Descriptor.ID] = append(r.dependents[reg.Descriptor.ID], id)
			}
		}
	}

	// Kahn's algorithm: inDegree[id] = number of providers (edges) that id depends on.
	inDegree := make(map[pluginapi.PluginID]int)
	for id := range r.byID {
		inDegree[id] = 0
	}
	for id, caps := range r.dependencies {
		for _, cap := range caps {
			inDegree[id] += len(r.byCapability[cap])
		}
	}

	queue := make([]pluginapi.PluginID, 0, len(inDegree))
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []pluginapi.PluginID
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, depID := range r.dependents[id] {
			inDegree[depID]--
			if inDegree[depID] == 0 {
				queue = append(queue, depID)
			}
		}
	}

	if len(order) != len(r.byID) {
		return ErrDependencyCycle
	}
	r.sortedStartup = order
	return nil
}

// StartupOrder returns the plugin IDs in dependency-respecting startup order.
// Call BuildDependencyGraph first. Returns empty if BuildDependencyGraph was not called.
func (r *Registry) StartupOrder() []pluginapi.PluginID {
	out := make([]pluginapi.PluginID, len(r.sortedStartup))
	copy(out, r.sortedStartup)
	return out
}

// AllPluginIDs returns all registered plugin IDs in no particular order.
func (r *Registry) AllPluginIDs() []pluginapi.PluginID {
	ids := make([]pluginapi.PluginID, 0, len(r.byID))
	for id := range r.byID {
		ids = append(ids, id)
	}
	return ids
}

// RegistrationFor returns the registration for a given ID, if present.
func (r *Registry) RegistrationFor(id pluginapi.PluginID) (*Registration, bool) {
	reg, ok := r.byID[id]
	return reg, ok
}

