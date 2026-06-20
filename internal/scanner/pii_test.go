package scanner

import (
	"sort"
	"testing"
)

func TestScanDocuments_email(t *testing.T) {
	docs := []string{"Contact us at support@example.com for help."}
	sigs := scanDocuments(docs, nil)
	if !contains(sigs, "email") {
		t.Errorf("expected email signal, got %v", sigs)
	}
}

func TestScanDocuments_apikey(t *testing.T) {
	docs := []string{"Token: sk-abcdefghijklmnopqrstuvwxyz1234"}
	sigs := scanDocuments(docs, nil)
	if !contains(sigs, "api_key") {
		t.Errorf("expected api_key signal, got %v", sigs)
	}
}

func TestScanDocuments_medical(t *testing.T) {
	docs := []string{"Patient admitted with diagnosis of pneumonia."}
	sigs := scanDocuments(docs, nil)
	if !contains(sigs, "medical") {
		t.Errorf("expected medical signal, got %v", sigs)
	}
}

func TestScanDocuments_personalNameInMeta(t *testing.T) {
	metas := []map[string]interface{}{
		{"author": "Jane Doe", "topic": "quarterly report"},
	}
	sigs := scanDocuments(nil, metas)
	if !contains(sigs, "personal_name") {
		t.Errorf("expected personal_name signal, got %v", sigs)
	}
}

func TestScanDocuments_clean(t *testing.T) {
	docs := []string{"This is a generic technical document with no sensitive content."}
	sigs := scanDocuments(docs, nil)
	if len(sigs) != 0 {
		t.Errorf("expected no signals, got %v", sigs)
	}
}

func TestMetadataKeys(t *testing.T) {
	metas := []map[string]interface{}{
		{"author": "Alice", "topic": "AI"},
		{"author": "Bob", "year": "2025"},
	}
	keys := metadataKeys(metas)
	sort.Strings(keys)
	want := []string{"author", "topic", "year"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i, k := range keys {
		if k != want[i] {
			t.Errorf("keys[%d] = %q, want %q", i, k, want[i])
		}
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
