package main

import (
	"fmt"
	"os"
	"path/filepath"

	agentconfig "github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	proxyconfig "github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	"github.com/ArmanAvanesyan/go-config/schema"
)

func main() {
	if err := generateSchemas(); err != nil {
		fmt.Fprintf(os.Stderr, "schema generation failed: %v\n", err)
		os.Exit(1)
	}
}

func generateSchemas() error {
	const dir = "schemas"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create schemas dir: %w", err)
	}

	if err := writeSchema(filepath.Join(dir, "agent.schema.json"), "AuthSentinel Agent config", schema.GenerateFor[agentconfig.Config]); err != nil {
		return err
	}
	if err := writeSchema(filepath.Join(dir, "proxy.schema.json"), "AuthSentinel Proxy config", schema.GenerateFor[proxyconfig.Config]); err != nil {
		return err
	}

	return nil
}

func writeSchema(path, title string, gen func(opts ...schema.Option) ([]byte, error)) error {
	b, err := gen(schema.WithTitle(title))
	if err != nil {
		return fmt.Errorf("generate schema for %s: %w", title, err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

