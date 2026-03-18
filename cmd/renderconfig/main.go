// renderconfig prints an example agent or proxy config with ApplyDefaults applied
// (JSON or YAML). Use for make render-config-example.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	agentconfig "github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	proxyconfig "github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	"gopkg.in/yaml.v3"
)

func main() {
	binary := os.Getenv("BINARY")
	if binary == "" {
		binary = "agent"
	}
	format := strings.ToLower(os.Getenv("FORMAT"))
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "yaml" {
		fmt.Fprintf(os.Stderr, "FORMAT must be json or yaml\n")
		os.Exit(2)
	}

	if err := run(binary, format); err != nil {
		fmt.Fprintf(os.Stderr, "render-config-example: %v\n", err)
		os.Exit(1)
	}
}

func run(binary, format string) error {
	switch binary {
	case "agent":
		var cfg agentconfig.Config
		cfg.ApplyDefaults()
		return emit(&cfg, format)
	case "proxy":
		var cfg proxyconfig.Config
		cfg.ApplyDefaults()
		return emit(&cfg, format)
	default:
		return fmt.Errorf("unknown BINARY=%q (use agent or proxy)", binary)
	}
}

func emit(cfg any, format string) error {
	if format == "yaml" {
		raw, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}
		out, err := yaml.Marshal(m)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(out)
		return err
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
