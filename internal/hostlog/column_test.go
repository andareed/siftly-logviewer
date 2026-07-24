package hostlog

import "testing"

func TestHostlogSchemaWrapsDetailsOnly(t *testing.T) {
	schema := hostlogColumnSchema()
	if got := schema.RoleDefaults[RolePrimary].WrapLines; got != 4 {
		t.Fatalf("Details WrapLines=%d want 4", got)
	}
	if got := schema.RoleDefaults[RoleSecondary].WrapLines; got != 0 {
		t.Fatalf("secondary WrapLines=%d want dense", got)
	}
}
