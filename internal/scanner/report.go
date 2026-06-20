package scanner

import (
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// PrintReport writes a human-readable report to stdout.
func PrintReport(r *ScanResult) {
	fmt.Printf("\n%s%s chromascan :: %s%s\n", colorBold, colorCyan, r.Target, colorReset)
	fmt.Printf("%s%s%s\n\n", colorDim, r.ScanTime.UTC().Format("2006-01-02T15:04:05Z"), colorReset)

	// Meta
	authStr := fmt.Sprintf("%sOPEN (no auth)%s", colorRed, colorReset)
	if r.AuthStatus == "AUTH_REQUIRED" {
		authStr = fmt.Sprintf("%sAUTH REQUIRED%s", colorGreen, colorReset)
	}
	fmt.Printf("## Meta\n")
	fmt.Printf("  version    : %s\n", r.Version)
	fmt.Printf("  api        : %s\n", r.APIVersion)
	fmt.Printf("  auth       : %s\n\n", authStr)

	// Collections
	if len(r.Collections) > 0 {
		fmt.Printf("## Collections (%d)\n", len(r.Collections))
		fmt.Printf("  %-32s  %8s  %5s  %s\n", "COLLECTION", "COUNT", "SCORE", "SEVERITY")
		fmt.Printf("  %s\n", strings.Repeat("-", 66))
		for _, cr := range r.Collections {
			scoreColor := colorReset
			switch cr.Severity {
			case "CRITICAL":
				scoreColor = colorRed
			case "HIGH":
				scoreColor = colorRed
			case "MEDIUM":
				scoreColor = colorYellow
			}
			piiTag := ""
			if len(cr.PIISignals) > 0 {
				piiTag = fmt.Sprintf(" [PII: %s]", strings.Join(cr.PIISignals, ","))
			}
			fmt.Printf("  %-32s  %8d  %s%5.1f%s  %s%s\n",
				cr.Name,
				cr.RecordCount,
				scoreColor, cr.Score, colorReset,
				cr.Severity,
				piiTag,
			)
		}
		fmt.Println()
	}

	// Sample documents
	for _, cr := range r.Collections {
		if len(cr.SampleDocuments) == 0 {
			continue
		}
		fmt.Printf("## Samples: %s\n", cr.Name)
		for i, doc := range cr.SampleDocuments {
			fmt.Printf("  [%d] %s\n", i+1, truncate(doc, 100))
		}
		if len(cr.MetadataKeys) > 0 {
			fmt.Printf("  metadata keys: %s\n", strings.Join(cr.MetadataKeys, ", "))
		}
		fmt.Println()
	}

	// Canary
	if r.Canary != nil && r.Canary.Attempted {
		fmt.Printf("## Write Canary\n")
		if r.Canary.Success {
			fmt.Printf("  %s[!] WRITE ACCESS CONFIRMED  canary-id=%s%s\n",
				colorRed, r.Canary.CanaryID, colorReset)
		} else {
			fmt.Printf("  write test failed: %s\n", r.Canary.Error)
		}
		fmt.Println()
	}

	// Summary
	sevColor := colorReset
	switch r.Severity {
	case "CRITICAL", "HIGH":
		sevColor = colorRed
	case "MEDIUM":
		sevColor = colorYellow
	}

	fmt.Printf("## Summary\n")
	fmt.Printf("  severity   : %s%s%s%s\n", colorBold, sevColor, r.Severity, colorReset)
	fmt.Printf("  score      : %.1f\n", r.Score)
	fmt.Printf("  auth       : %s\n", r.AuthStatus)
	fmt.Printf("  collections: %d total\n", r.Summary.TotalCollections)
	fmt.Printf("  objects    : ~%d estimated\n", r.Summary.TotalRecordsEstimated)
	fmt.Printf("  pii hits   : %d collections\n", r.Summary.PIICollections)
	fmt.Printf("  findings   : %d extracted\n", r.Summary.FindingsExtracted)
	fmt.Println()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
