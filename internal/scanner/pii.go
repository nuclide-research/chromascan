package scanner

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reEmail  = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	reAPIKey = regexp.MustCompile(`(sk-|AKIA|Bearer |ghp_)[a-zA-Z0-9]{20,}`)
)

// medicalTerms are checked as case-insensitive substrings.
var medicalTerms = []string{
	"patient", "diagnosis", "prescription", "clinical", "phi",
}

// nameFields are metadata field names that often carry personal names.
var nameFields = []string{"author", "user", "name", "email"}

// piiSignal is a short tag describing what was found.
type piiSignal string

const (
	sigEmail   piiSignal = "email"
	sigAPIKey  piiSignal = "api_key"
	sigMedical piiSignal = "medical"
	sigName    piiSignal = "personal_name"
)

// scanDocuments checks document texts and metadata values for PII signals.
// Returns a deduplicated slice of signal tags.
func scanDocuments(docs []string, metas []map[string]interface{}) []string {
	found := make(map[piiSignal]bool)

	for _, doc := range docs {
		if reEmail.MatchString(doc) {
			found[sigEmail] = true
		}
		if reAPIKey.MatchString(doc) {
			found[sigAPIKey] = true
		}
		lower := strings.ToLower(doc)
		for _, term := range medicalTerms {
			if strings.Contains(lower, term) {
				found[sigMedical] = true
				break
			}
		}
	}

	for _, meta := range metas {
		for k, v := range meta {
			kl := strings.ToLower(k)
			val := fmt.Sprintf("%v", v)

			// Check field name against personal name fields.
			for _, nf := range nameFields {
				if strings.Contains(kl, nf) && len(val) > 0 {
					found[sigName] = true
					break
				}
			}
			if reEmail.MatchString(val) {
				found[sigEmail] = true
			}
			if reAPIKey.MatchString(val) {
				found[sigAPIKey] = true
			}
		}
	}

	var out []string
	for sig := range found {
		out = append(out, string(sig))
	}
	return out
}

// metadataKeys returns the unique set of keys across all metadata maps.
func metadataKeys(metas []map[string]interface{}) []string {
	seen := make(map[string]bool)
	for _, m := range metas {
		for k := range m {
			seen[k] = true
		}
	}
	var keys []string
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}
