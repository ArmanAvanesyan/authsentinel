package builtin

import (
	"context"

	"github.com/ArmanAvanesyan/authsentinel/pkg/plugindiscovery"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginregistry"
	"github.com/ArmanAvanesyan/authsentinel/pkg/plugins/pipeline/ratelimit"
	provideroidc "github.com/ArmanAvanesyan/authsentinel/pkg/plugins/provider/oidc"
)

// Registrar registers built-in plugins that are compiled into the binary.
// It is used by cmd/* before manifest discovery and before runtime execution.
type Registrar struct{}

var _ plugindiscovery.BuiltinRegistrar = (*Registrar)(nil)

func (r *Registrar) RegisterBuiltins(ctx context.Context, reg *pluginregistry.Registry) error {
	// Pipeline: rate limit
	p := ratelimit.New()
	desc := p.Descriptor()
	if err := reg.Register(desc, func(ctx context.Context, _ pluginapi.PluginDescriptor) (pluginapi.Plugin, error) {
		return ratelimit.New(), nil
	}); err != nil {
		return err
	}

	// Provider: OIDC
	po := provideroidc.New()
	pdesc := po.Descriptor()
	return reg.Register(pdesc, func(ctx context.Context, _ pluginapi.PluginDescriptor) (pluginapi.Plugin, error) {
		return provideroidc.New(), nil
	})
}

