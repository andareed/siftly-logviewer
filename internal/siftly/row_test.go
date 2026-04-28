package siftly

import (
	"strings"
	"testing"
)

func legacyComputeID(cols []string) uint64 {
	var h uint64 = 14695981039346656037
	for _, col := range cols {
		norm := strings.ToLower(strings.TrimSpace(col))
		for i := 0; i < len(norm); i++ {
			h ^= uint64(norm[i])
			h *= 1099511628211
		}
		h ^= 0
		h *= 1099511628211
	}
	return h
}

func TestComputeIDMatchesLegacyNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cols []string
	}{
		{
			name: "ascii trim and lower",
			cols: []string{"  Alpha\t", "Beta ", "\nGamma\r"},
		},
		{
			name: "empty and whitespace only",
			cols: []string{"", "   ", "\t"},
		},
		{
			name: "unicode falls back exactly",
			cols: []string{"  Café  ", "\u00a0MiXeD\u00a0", "Straße"},
		},
		{
			name: "mixed ascii and unicode",
			cols: []string{"Host-1", " Δelta ", "VALUE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := Row{Cols: tt.cols}
			got := row.ComputeID()
			want := legacyComputeID(tt.cols)
			if got != want {
				t.Fatalf("ComputeID() = %d want %d for %#v", got, want, tt.cols)
			}
		})
	}
}
