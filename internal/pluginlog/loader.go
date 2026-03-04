package pluginlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

func LoadModelAuto(path string) (*siftly.Model, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return newModelFromJSONFile(path)
	default:
		return newModelFromLogFile(path)
	}
}

func newModelFromJSONFile(path string) (*siftly.Model, error) {
	m := &siftly.Model{}
	if err := siftly.LoadModel(m, path); err != nil {
		return nil, err
	}
	m.InitialPath = path
	m.SetStyles(SiftlyStyles())
	m.InitialiseView()
	return m, nil
}

func newModelFromLogFile(path string) (*siftly.Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	records, err := parsePluginLog(string(data))
	if err != nil {
		return nil, err
	}

	m, err := siftly.NewModelFromRecords(records, pluginlogColumnSchema())
	if err != nil {
		return nil, err
	}
	m.InitialPath = path
	m.SetStyles(SiftlyStyles())
	m.InitialiseView()
	return m, nil
}
