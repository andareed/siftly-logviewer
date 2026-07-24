package todaylog

import "testing"

func TestTodaylogColumnLayoutPrioritisesKey(t *testing.T) {
	schema := todaylogColumnSchema()
	key := schema.ColumnDefaults["key"]
	value := schema.ColumnDefaults["value"]

	if got := detectRole("key"); got != RolePrimary {
		t.Fatalf("key role = %v, want primary", got)
	}
	if got := detectRole("value"); got != RoleSecondary {
		t.Fatalf("value role = %v, want secondary", got)
	}
	if key.MinWidth < 40 || key.Weight <= value.Weight {
		t.Fatalf("key layout must dominate value: key=%+v value=%+v", key, value)
	}
	if key.WrapLines != 4 || value.WrapLines != 0 {
		t.Fatalf("key should wrap while value remains dense: key=%+v value=%+v", key, value)
	}

	minWidth := 0
	for _, name := range todaylogHeader {
		minWidth += schema.ColumnDefaults[name].MinWidth
	}
	if minWidth > 100 {
		t.Fatalf("todaylog minimum widths total %d, want at most 100", minWidth)
	}
}
