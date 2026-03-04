package devfmt

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/andareed/siftly-hostlog/internal/siftly"
)

type Query struct {
	Group    string
	Category string
	ID       string
	Search   string
	SortID   bool
}

type FieldRow struct {
	GroupCategory string
	RawCategory   string
	ID            string
	FieldName     string
	FieldLabel    string
	RawValue      string
	DisplayValue  string
	Desc          string
	Time          string
	AgentID       string
	Extra         string
}

func LoadDataset(inputPath string) (*Dataset, MappingConfig, error) {
	groupCfg, err := LoadGroupConfig()
	if err != nil {
		return nil, nil, err
	}
	mappingCfg, err := LoadMappingConfig()
	if err != nil {
		return nil, nil, err
	}
	norm := NewGroupNormalizer(groupCfg)
	ds := NewDataset()

	r, closer, err := openInput(inputPath)
	if err != nil {
		return nil, nil, err
	}
	if closer != nil {
		defer closer.Close()
	}

	if err := ParseDump(r, func(row ParsedRow) error {
		if row.FieldValueLC == "" && row.FieldValue != "" {
			row.FieldValueLC = strings.ToLower(row.FieldValue)
		}
		return ds.addRow(row, norm)
	}); err != nil {
		return nil, nil, err
	}

	return ds, mappingCfg, nil
}

func BuildRows(ds *Dataset, mappings MappingConfig, q Query) []FieldRow {
	rows := make([]FieldRow, 0, len(ds.Entities)*4)
	search := strings.ToLower(strings.TrimSpace(q.Search))
	for _, ent := range filteredEntities(ds.Entities, q) {
		for _, fname := range ent.FieldOrder {
			fv := ent.Fields[fname]
			if search != "" && !strings.Contains(fv.RawValueLC, search) {
				continue
			}

			fm, ok := mappings.Lookup(ent.GroupCategory, ent.RawCategory, fname)
			if ok && fm.Hide {
				continue
			}

			label := fname
			if ok && fm.Label != "" {
				label = fm.Label
			}
			decoded := decodeValue(fv.RawValue, fv.RawValueLC, fm, ok)

			rows = append(rows, FieldRow{
				GroupCategory: ent.GroupCategory,
				RawCategory:   ent.RawCategory,
				ID:            ent.ID,
				FieldName:     fname,
				FieldLabel:    label,
				RawValue:      fv.RawValue,
				DisplayValue:  decoded,
				Desc:          fm.Desc,
				Time:          ent.Time,
				AgentID:       ent.AgentID,
				Extra:         ent.Extra,
			})
		}
	}
	return rows
}

func BuildRecords(rows []FieldRow) [][]string {
	records := [][]string{{
		"group_category",
		"category",
		"id",
		"field_name",
		"field_label",
		"value",
		"raw_value",
		"time",
		"agentid",
		"extra",
		"desc",
	}}

	for _, row := range rows {
		records = append(records, []string{
			row.GroupCategory,
			row.RawCategory,
			row.ID,
			row.FieldName,
			row.FieldLabel,
			row.DisplayValue,
			row.RawValue,
			row.Time,
			row.AgentID,
			row.Extra,
			row.Desc,
		})
	}
	return records
}

func BuildModel(inputPath string, rows []FieldRow) (*siftly.Model, error) {
	records := BuildRecords(rows)
	m, err := siftly.NewModelFromRecords(records, devfmtColumnSchema())
	if err != nil {
		return nil, err
	}
	m.InitialPath = inputPath
	m.SetStyles(SiftlyStyles())
	m.InitialiseView()
	return m, nil
}

func GroupsSeen(ds *Dataset) []string {
	out := make([]string, len(ds.GroupsSeenOrder))
	copy(out, ds.GroupsSeenOrder)
	return out
}

func CategoriesSeen(ds *Dataset) []string {
	out := make([]string, len(ds.CategoriesSeenOrder))
	copy(out, ds.CategoriesSeenOrder)
	return out
}

func filteredEntities(entities []*Entity, q Query) []*Entity {
	out := make([]*Entity, 0, len(entities))
	for _, ent := range entities {
		if q.Group != "" && ent.GroupCategory != q.Group {
			continue
		}
		if q.Category != "" && ent.RawCategory != q.Category {
			continue
		}
		if q.ID != "" && ent.ID != q.ID {
			continue
		}
		out = append(out, ent)
	}
	if q.SortID {
		sort.SliceStable(out, func(i, j int) bool {
			if out[i].GroupCategory != out[j].GroupCategory {
				return out[i].GroupCategory < out[j].GroupCategory
			}
			if out[i].ID != out[j].ID {
				return out[i].ID < out[j].ID
			}
			return out[i].SeenOrder < out[j].SeenOrder
		})
	}
	return out
}

func decodeValue(raw string, rawLC string, fm FieldMapping, hasMapping bool) string {
	if !hasMapping {
		return raw
	}
	value := raw
	if fm.BoolNormalize {
		if b, ok := normalizeBool(rawLC); ok {
			if b {
				value = "true"
			} else {
				value = "false"
			}
		}
	}
	if len(fm.Enum) > 0 {
		if ev, ok := fm.Enum[raw]; ok {
			value = ev
		} else if ev, ok := fm.Enum[rawLC]; ok {
			value = ev
		}
	}
	if fm.Unit != "" && value != "" {
		value = value + " " + fm.Unit
	}
	if fm.Format != "" {
		value = fmt.Sprintf(fm.Format, value)
	}
	return value
}

func normalizeBool(v string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func openInput(path string) (io.Reader, io.Closer, error) {
	if path == "" || path == "-" {
		return os.Stdin, nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open %q: %w", path, err)
	}
	return f, f, nil
}
