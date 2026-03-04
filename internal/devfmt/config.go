package devfmt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var defaultGroupConfig = GroupConfig{
	Nosplit: []string{},
	Exact:   map[string]string{},
	Prefix: map[string]string{
		"sw_": "sw",
	},
}

var defaultMappingConfig = MappingConfig{}

func LoadGroupConfig() (GroupConfig, error) {
	cfg := defaultGroupConfig
	if err := loadConfigFile("groups", &cfg); err != nil {
		return GroupConfig{}, err
	}
	if cfg.Exact == nil {
		cfg.Exact = map[string]string{}
	}
	if cfg.Prefix == nil {
		cfg.Prefix = map[string]string{}
	}
	return cfg, nil
}

func LoadMappingConfig() (MappingConfig, error) {
	cfg := defaultMappingConfig
	if err := loadConfigFile("mappings", &cfg); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = MappingConfig{}
	}
	return cfg, nil
}

func loadConfigFile(base string, out any) error {
	candidates := []string{
		filepath.Join(".", base+".yaml"),
		filepath.Join(".", base+".yml"),
		filepath.Join(".", base+".json"),
	}
	var found string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			found = p
			break
		}
	}
	if found == "" {
		return nil
	}

	b, err := os.ReadFile(found)
	if err != nil {
		return fmt.Errorf("read %s: %w", found, err)
	}
	if strings.HasSuffix(found, ".json") {
		if err := json.Unmarshal(b, out); err != nil {
			return fmt.Errorf("parse %s: %w", found, err)
		}
		return nil
	}
	if err := yaml.Unmarshal(b, out); err != nil {
		var te *yaml.TypeError
		if errors.As(err, &te) {
			return fmt.Errorf("parse %s: %s", found, strings.Join(te.Errors, "; "))
		}
		return fmt.Errorf("parse %s: %w", found, err)
	}
	return nil
}
