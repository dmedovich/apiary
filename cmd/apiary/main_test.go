package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestChooseFormat(t *testing.T) {
	cases := []struct {
		flagFmt, cfgFmt, out, want string
	}{
		{"json", "", "x.yaml", "json"},
		{"", "json", "x.yaml", "json"},
		{"", "", "x.json", "json"},
		{"", "", "x.JSON", "json"},
		{"", "", "openapi.yaml", "yaml"},
		{"yaml", "json", "x.json", "yaml"},
	}
	for _, c := range cases {
		if got := chooseFormat(c.flagFmt, c.cfgFmt, c.out); got != c.want {
			t.Errorf("chooseFormat(%q,%q,%q) = %q, want %q", c.flagFmt, c.cfgFmt, c.out, got, c.want)
		}
	}
}

func TestSplitCSV(t *testing.T) {
	got := splitCSV(" a , b ,, c ")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
	if splitCSV("") != nil {
		t.Error("empty string should yield nil")
	}
}

func TestResolveStr(t *testing.T) {
	if got := resolveStr(true, "flag", "cfg"); got != "flag" {
		t.Errorf("set flag should win, got %q", got)
	}
	if got := resolveStr(false, "default", "cfg"); got != "cfg" {
		t.Errorf("config should win when flag unset, got %q", got)
	}
	if got := resolveStr(false, "default", ""); got != "default" {
		t.Errorf("default should be used, got %q", got)
	}
}

func TestEncodeSpec_YAMLvsJSON(t *testing.T) {
	spec, _, err := buildSpec([]string{"../../testdata/sample"}, specOptions{title: "T", version: "1"})
	if err != nil {
		t.Fatal(err)
	}

	y, err := encodeSpec(spec, "yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(y), "openapi: 3.1.0") {
		t.Errorf("yaml output unexpected start: %.20q", y)
	}

	j, err := encodeSpec(spec, "json")
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(j, &v); err != nil {
		t.Fatalf("json output is not valid JSON: %v", err)
	}
	if v["openapi"] != "3.1.0" {
		t.Errorf("json openapi field = %v", v["openapi"])
	}
}
