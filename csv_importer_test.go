package uadmin

import (
	"testing"
)

func Test_getModelDataMapping(t *testing.T) {
	rows := []string{}
	_, err := getModelDataMapping(rows)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	rows = []string{"1;en"}
	_, err = getModelDataMapping(rows)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	rows = []string{"1;en;test;fields"}
	entries, err := getModelDataMapping(rows)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].LangFields["en"][0] != "test" {
		t.Errorf("expected 'test', got %s", entries[0].LangFields["en"][0])
	}
	if entries[0].LangFields["en"][1] != "fields" {
		t.Errorf("expected 'fields', got %s", entries[0].LangFields["en"][1])
	}
}
