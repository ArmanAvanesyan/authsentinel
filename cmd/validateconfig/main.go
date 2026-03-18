// validateconfig loads an agent or proxy config file and runs the same validation
// as the runtime (ApplyDefaults + Validate). Use for make validate-config.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	agentconfig "github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	proxyconfig "github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	goconfig "github.com/ArmanAvanesyan/go-config/config"
	jsonformat "github.com/ArmanAvanesyan/go-config/format/json"
	"github.com/ArmanAvanesyan/go-config/source/file"
	"gopkg.in/yaml.v3"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = os.Getenv("AGENT_CONFIG")
	}
	if configPath == "" {
		configPath = os.Getenv("PROXY_CONFIG")
	}
	binary := os.Getenv("BINARY")
	if binary == "" && configPath != "" {
		if strings.Contains(configPath, "proxy") {
			binary = "proxy"
		} else {
			binary = "agent"
		}
	}
	if binary == "" {
		binary = "agent"
	}

	if configPath == "" {
		fmt.Fprintf(os.Stderr, "usage: CONFIG_PATH=/path/to/config.json BINARY=agent|proxy %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  or set AGENT_CONFIG / PROXY_CONFIG instead of CONFIG_PATH\n")
		os.Exit(2)
	}

	if err := run(configPath, binary); err != nil {
		fmt.Fprintf(os.Stderr, "validate-config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("config valid")
}

func run(configPath, binary string) error {
	ext := strings.ToLower(filepath.Ext(configPath))
	isYAML := ext == ".yaml" || ext == ".yml"

	if isYAML {
		return validateYAML(configPath, binary)
	}
	return validateWithGoConfig(configPath, binary)
}

func validateYAML(configPath, binary string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parse YAML: %w", err)
	}
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("convert to JSON: %w", err)
	}
	switch binary {
	case "agent":
		var cfg agentconfig.Config
		if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
			return fmt.Errorf("decode config: %w", err)
		}
		cfg.ApplyDefaults()
		return cfg.Validate()
	case "proxy":
		var cfg proxyconfig.Config
		if err := json.Unmarshal(jsonBytes, &cfg); err != nil {
			return fmt.Errorf("decode config: %w", err)
		}
		cfg.ApplyDefaults()
		return cfg.Validate()
	default:
		return fmt.Errorf("unknown BINARY=%q (use agent or proxy)", binary)
	}
}

func validateWithGoConfig(configPath, binary string) error {
	ctx := context.Background()
	loader := goconfig.New().AddSource(file.New(configPath), jsonformat.New())

	switch binary {
	case "agent":
		var cfg agentconfig.Config
		if err := loader.Load(ctx, &cfg); err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfg.ApplyDefaults()
		return cfg.Validate()
	case "proxy":
		var cfg proxyconfig.Config
		if err := loader.Load(ctx, &cfg); err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		cfg.ApplyDefaults()
		return cfg.Validate()
	default:
		return fmt.Errorf("unknown BINARY=%q (use agent or proxy)", binary)
	}
}
