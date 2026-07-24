package pluginlog

import "testing"

func TestPluginlogSchemaWrapsMessageOnly(t *testing.T) {
	schema := pluginlogColumnSchema()
	if got := schema.ColumnDefaults["message"].WrapLines; got != 4 {
		t.Fatalf("Message WrapLines=%d want 4", got)
	}
	if got := schema.ColumnDefaults["process"].WrapLines; got != 0 {
		t.Fatalf("Process WrapLines=%d want dense", got)
	}
}
