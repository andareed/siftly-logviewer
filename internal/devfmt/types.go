package devfmt

import (
	"sort"
	"strings"
)

type FieldValue struct {
	RawValue   string
	RawValueLC string
}

type Entity struct {
	RawCategory   string
	GroupCategory string
	ID            string
	AgentID       string
	Time          string
	Extra         string
	Fields        map[string]FieldValue
	FieldOrder    []string
	SeenOrder     int
}

type Dataset struct {
	Entities            []*Entity
	entityIndex         map[string]*Entity
	GroupsSeenOrder     []string
	CategoriesSeenOrder []string
	groupsSet           map[string]struct{}
	categoriesSet       map[string]struct{}
}

func NewDataset() *Dataset {
	return &Dataset{
		Entities:      make([]*Entity, 0, 1024),
		entityIndex:   make(map[string]*Entity, 1024),
		groupsSet:     make(map[string]struct{}, 128),
		categoriesSet: make(map[string]struct{}, 128),
	}
}

func (d *Dataset) addRow(row ParsedRow, gnorm *GroupNormalizer) error {
	if row.Category == "" || row.ID == "" {
		return nil
	}
	group := gnorm.Normalize(row.Category)
	if _, ok := d.groupsSet[group]; !ok {
		d.groupsSet[group] = struct{}{}
		d.GroupsSeenOrder = append(d.GroupsSeenOrder, group)
	}
	if _, ok := d.categoriesSet[row.Category]; !ok {
		d.categoriesSet[row.Category] = struct{}{}
		d.CategoriesSeenOrder = append(d.CategoriesSeenOrder, row.Category)
	}

	key := row.Category + "\x00" + row.ID
	ent, ok := d.entityIndex[key]
	if !ok {
		ent = &Entity{
			RawCategory:   row.Category,
			GroupCategory: group,
			ID:            row.ID,
			AgentID:       row.AgentID,
			Time:          row.Time,
			Extra:         row.Extra,
			Fields:        make(map[string]FieldValue, 32),
			FieldOrder:    make([]string, 0, 32),
			SeenOrder:     len(d.Entities),
		}
		d.entityIndex[key] = ent
		d.Entities = append(d.Entities, ent)
	} else {
		if ent.AgentID == "" {
			ent.AgentID = row.AgentID
		}
		if ent.Time == "" {
			ent.Time = row.Time
		}
		if ent.Extra == "" {
			ent.Extra = row.Extra
		}
	}

	if row.FieldName != "" {
		if _, exists := ent.Fields[row.FieldName]; !exists {
			ent.Fields[row.FieldName] = FieldValue{RawValue: row.FieldValue, RawValueLC: row.FieldValueLC}
			ent.FieldOrder = append(ent.FieldOrder, row.FieldName)
		}
	}

	return nil
}

type GroupConfig struct {
	Nosplit []string          `yaml:"nosplit" json:"nosplit"`
	Exact   map[string]string `yaml:"exact" json:"exact"`
	Prefix  map[string]string `yaml:"prefix" json:"prefix"`
}

type prefixRule struct {
	Prefix string
	Group  string
}

type GroupNormalizer struct {
	nosplit map[string]struct{}
	exact   map[string]string
	prefix  []prefixRule
}

func NewGroupNormalizer(cfg GroupConfig) *GroupNormalizer {
	gn := &GroupNormalizer{
		nosplit: make(map[string]struct{}, len(cfg.Nosplit)),
		exact:   make(map[string]string, len(cfg.Exact)),
		prefix:  make([]prefixRule, 0, len(cfg.Prefix)),
	}
	for _, c := range cfg.Nosplit {
		gn.nosplit[strings.TrimSpace(c)] = struct{}{}
	}
	for k, v := range cfg.Exact {
		gn.exact[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	for p, g := range cfg.Prefix {
		gn.prefix = append(gn.prefix, prefixRule{Prefix: strings.TrimSpace(p), Group: strings.TrimSpace(g)})
	}
	sort.Slice(gn.prefix, func(i, j int) bool {
		if len(gn.prefix[i].Prefix) != len(gn.prefix[j].Prefix) {
			return len(gn.prefix[i].Prefix) > len(gn.prefix[j].Prefix)
		}
		return gn.prefix[i].Prefix < gn.prefix[j].Prefix
	})
	return gn
}

func (gn *GroupNormalizer) Normalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if _, ok := gn.nosplit[raw]; ok {
		return raw
	}
	if g, ok := gn.exact[raw]; ok && g != "" {
		return g
	}
	for _, rule := range gn.prefix {
		if rule.Prefix != "" && strings.HasPrefix(raw, rule.Prefix) {
			return rule.Group
		}
	}
	if i := strings.IndexByte(raw, '_'); i > 0 {
		return raw[:i]
	}
	return raw
}

type FieldMapping struct {
	Label         string            `yaml:"label" json:"label"`
	Desc          string            `yaml:"desc" json:"desc"`
	Hide          bool              `yaml:"hide" json:"hide"`
	Enum          map[string]string `yaml:"enum" json:"enum"`
	Unit          string            `yaml:"unit" json:"unit"`
	Format        string            `yaml:"format" json:"format"`
	BoolNormalize bool              `yaml:"bool_normalize" json:"bool_normalize"`
}

type MappingConfig map[string]map[string]FieldMapping

func (m MappingConfig) Lookup(groupCategory, rawCategory, fieldName string) (FieldMapping, bool) {
	if fields, ok := m[groupCategory]; ok {
		if fm, ok := fields[fieldName]; ok {
			return fm, true
		}
	}
	if fields, ok := m[rawCategory]; ok {
		if fm, ok := fields[fieldName]; ok {
			return fm, true
		}
	}
	return FieldMapping{}, false
}
