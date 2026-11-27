package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestGeographyValidateOnlyNeedsNoDatabase(t *testing.T) {
	var output bytes.Buffer
	err := run([]string{"-file", "../../../../data/iran-geography/iran-provinces-cities.json", "-validate-only"}, func(string) string { t.Fatal("offline validation accessed the environment"); return "" }, &output)
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(output.Bytes(), &report); err != nil || report["applied"] != false || report["source_cities"] != float64(1481) || report["validation"] != "passed" {
		t.Fatal("invalid validation report", err)
	}
}

func TestGeographyImporterRequiresAnExplicitMatchingDatabase(t *testing.T) {
	for _, test := range []struct {
		mode, live, review string
		valid              bool
	}{
		{"review", "", "review-db", true}, {"live", "live-db", "", true},
		{"review", "live-db", "", false}, {"live", "", "review-db", false},
		{"review", "live-db", "review-db", false}, {"review", "", "", false},
	} {
		getenv := func(key string) string {
			if key == "DATABASE_URL" {
				return test.live
			}
			return test.review
		}
		_, err := targetDSN(test.mode, getenv)
		if (err == nil) != test.valid {
			t.Fatalf("target guard mode=%s: %v", test.mode, err)
		}
	}
	for _, args := range [][]string{{"-file", "file.json", "-apply"}, {"-file", "file.json", "-mode", "other"}, {"-file", "file.json", "-apply", "-backup", "file.bak", "-validate-only"}} {
		if err := run(args, func(string) string { return "" }, &bytes.Buffer{}); err == nil {
			t.Fatal("invalid execution flags accepted")
		}
	}
}
