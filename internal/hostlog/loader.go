package hostlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/shared/logging"
	"github.com/andareed/siftly-hostlog/internal/siftly"
)

func LoadModelAuto(path string) (*siftly.Model, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return newModelFromJSONFile(path)
	case ".csv":
		return newModelFromCSVFile(path)
	default:
		return nil, fmt.Errorf("unsupported file extension %q (want .csv or .json)", ext)
	}
}

// Load Data From Serialized JSONs using LoadModel(m, path)
// Implies that this has been analysed previously and saved
func newModelFromJSONFile(path string) (*siftly.Model, error) {
	defer logging.TimeOperation("load hostlog json")()

	m := &siftly.Model{}
	if err := siftly.LoadModel(m, path); err != nil {
		return nil, err
	}
	m.InitialPath = path
	m.SetStyles(SiftlyStyles())
	m.InitialiseView()
	return m, nil
}

func newModelFromCSVFile(path string) (*siftly.Model, error) {
	defer logging.TimeOperation("load hostlog csv")()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	m, err := siftly.NewModelFromCSVReader(f, hostlogColumnSchema())
	if err != nil {
		return nil, err
	}
	m.InitialPath = path
	m.SetStyles(SiftlyStyles())
	m.InitialiseView()
	return m, nil
}
