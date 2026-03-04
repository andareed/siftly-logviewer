package pluginlog

import (
	"strings"
	"testing"
)

func TestParsePluginLog(t *testing.T) {
	content := strings.Join([]string{
		"sw:9320:1770790588.732206:Wed Feb 11 06:16:28 2026: Error: FSSwitchSNMP::switch_snmp_get_tab_internal:736: ip[10.2.0.154] error[No Transport Domain defined]",
		"sw:9320:1770790589.187061:Wed Feb 11 06:16:29 2026: main::sw_stats:13170:[]::0: Resolved[10] Blocked[0] Handles[0] Switches[4]",
	}, "\n")

	records, err := parsePluginLog(content)
	if err != nil {
		t.Fatalf("parsePluginLog returned error: %v", err)
	}
	if got, want := len(records), 3; got != want {
		t.Fatalf("record count=%d want=%d", got, want)
	}

	first := records[1]
	if first[0] != "sw" {
		t.Fatalf("process=%q want=%q", first[0], "sw")
	}
	if first[1] != "9320" {
		t.Fatalf("pid=%q want=%q", first[1], "9320")
	}
	if first[2] != "1770790588.732206" {
		t.Fatalf("epoch=%q want=%q", first[2], "1770790588.732206")
	}
	if first[3] != "2026-02-11 06:16:28" {
		t.Fatalf("time=%q want=%q", first[3], "2026-02-11 06:16:28")
	}
	if first[4] != "Error" {
		t.Fatalf("level=%q want=%q", first[4], "Error")
	}
	if first[5] != "FSSwitchSNMP" {
		t.Fatalf("module=%q want=%q", first[5], "FSSwitchSNMP")
	}
	if first[6] != "switch_snmp_get_tab_internal" {
		t.Fatalf("function=%q want=%q", first[6], "switch_snmp_get_tab_internal")
	}
	if first[7] != "736" {
		t.Fatalf("line=%q want=%q", first[7], "736")
	}
	if !strings.Contains(first[8], "No Transport Domain defined") {
		t.Fatalf("message=%q missing expected detail", first[8])
	}

	second := records[2]
	if second[4] != "" {
		t.Fatalf("second level=%q want empty", second[4])
	}
	if second[5] != "main" || second[6] != "sw_stats" || second[7] != "13170" {
		t.Fatalf("second mfl parsed unexpectedly: module=%q function=%q line=%q", second[5], second[6], second[7])
	}
	if !strings.HasPrefix(second[8], "[]::0:") {
		t.Fatalf("second message=%q expected payload prefix", second[8])
	}
}

func TestParsePluginLogMultilineAndSwShard(t *testing.T) {
	content := strings.Join([]string{
		"sw-12:345:1770790853.147421:Wed Feb 11 06:20:53 2026: Error: FSExpectCLI::handle_config:962: ip[10.2.0.50] internal_err before_match[show run interface",
		"% Incomplete command.",
		"",
		"] -- cmd[show run interface] error[]",
		"sw:9320:1770790853.147625:Wed Feb 11 06:20:53 2026: Error: FSSwitch::switch_get_running_config_cb:2616: ip[10.2.0.50] error[Enable Privileged Password not configured or invalid/incomplete command.]",
	}, "\n")

	records, err := parsePluginLog(content)
	if err != nil {
		t.Fatalf("parsePluginLog returned error: %v", err)
	}
	if got, want := len(records), 3; got != want {
		t.Fatalf("record count=%d want=%d", got, want)
	}

	first := records[1]
	if first[0] != "sw-12" {
		t.Fatalf("process=%q want=%q", first[0], "sw-12")
	}
	if first[4] != "Error" || first[5] != "FSExpectCLI" || first[6] != "handle_config" || first[7] != "962" {
		t.Fatalf("unexpected parsed fields level=%q module=%q function=%q line=%q", first[4], first[5], first[6], first[7])
	}
	if !strings.Contains(first[8], "% Incomplete command.") {
		t.Fatalf("multiline message lost expected body: %q", first[8])
	}
}

func TestParsePluginLogNoRecords(t *testing.T) {
	if _, err := parsePluginLog("no plugin rows"); err == nil {
		t.Fatalf("expected error for missing plugin rows")
	}
}
