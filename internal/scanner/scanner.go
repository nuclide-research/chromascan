package scanner

import "fmt"

// Scanner orchestrates all scan phases.
type Scanner struct {
	cfg *Config
	c   *httpClient
}

// New constructs a Scanner from cfg.
func New(cfg *Config) *Scanner {
	return &Scanner{
		cfg: cfg,
		c:   newClient(cfg.Target, cfg.TimeoutSec, cfg.Verbose),
	}
}

// Run executes the scan pipeline and populates result in place.
func (s *Scanner) Run(result *ScanResult) error {
	cfg := s.cfg
	c := s.c
	tenant := cfg.Tenant
	database := cfg.Database

	fmt.Printf("[0] probe %s\n", cfg.Target)

	// Phase 0: fingerprint
	pr, err := probe(c)
	if err != nil {
		return fmt.Errorf("probe: %w", err)
	}
	result.Version = pr.version
	result.APIVersion = string(pr.api)
	result.AuthStatus = pr.authStatus

	if pr.authStatus == "AUTH_REQUIRED" {
		fmt.Println("    auth required -- stopping")
		buildSummary(result)
		return nil
	}

	if cfg.ProbeOnly {
		buildSummary(result)
		return nil
	}

	// Phase 1: list collections
	fmt.Println("[1] listing collections")
	rawCols, err := listCollections(c, pr.api, tenant, database)
	if err != nil {
		fmt.Printf("    collection list error: %v\n", err)
		buildSummary(result)
		return nil
	}
	fmt.Printf("    %d collections found\n", len(rawCols))

	authOpen := pr.authStatus == "OPEN"

	// Phase 2: enumerate each collection
	fmt.Printf("[2] enumerating %d collections\n", len(rawCols))
	for _, rc := range rawCols {
		cr := enumerateCollection(c, pr.api, tenant, database, rc, authOpen, cfg.Verbose)
		result.Collections = append(result.Collections, cr)
	}

	// Write canary
	if cfg.DoWriteCanary {
		fmt.Println("[W] canary write")
		canary := injectCanary(c, pr.api, tenant, database, rawCols)
		result.Canary = canary
		if canary.Success {
			result.WriteTest = "SUCCESS"
			// Re-score all collections to include write confirmation.
			for i := range result.Collections {
				result.Collections[i].Score = scoreCollection(
					authOpen,
					result.Collections[i].PIISignals,
					result.Collections[i].RecordCount,
					true,
				)
				result.Collections[i].Severity = severityLabel(result.Collections[i].Score)
			}
		} else {
			result.WriteTest = "FAILED"
		}
	}

	buildSummary(result)
	return nil
}

// enumerateCollection fetches count, samples documents, and scores one collection.
func enumerateCollection(c *httpClient, api apiVersion, tenant, database string, rc rawCollection, authOpen, verbose bool) CollectionResult {
	cr := CollectionResult{
		Name: rc.Name,
		ID:   rc.ID,
	}
	if rc.Dimension != nil {
		cr.Dimension = *rc.Dimension
	}

	// Record count
	cr.RecordCount = fetchCount(c, api, tenant, database, rc.ID)

	// Sample documents + metadata
	samples, err := sampleRecords(c, api, tenant, database, rc.ID, 3)
	if err != nil {
		if verbose {
			fmt.Printf("    sample %s: %v\n", rc.Name, err)
		}
	} else {
		cr.SampleDocuments = samples.Documents
		cr.SampleMetadatas = samples.Metadatas
		cr.MetadataKeys = metadataKeys(samples.Metadatas)
		cr.PIISignals = scanDocuments(samples.Documents, samples.Metadatas)
	}

	cr.Score = scoreCollection(authOpen, cr.PIISignals, cr.RecordCount, false)
	cr.Severity = severityLabel(cr.Score)
	return cr
}

// buildSummary rolls up collection-level data into the top-level summary.
func buildSummary(r *ScanResult) {
	r.Score = overallScore(r.Collections)
	r.Severity = severityLabel(r.Score)

	sum := &r.Summary
	sum.AuthStatus = r.AuthStatus
	sum.TotalCollections = len(r.Collections)
	for _, cr := range r.Collections {
		sum.TotalRecordsEstimated += cr.RecordCount
		if len(cr.PIISignals) > 0 {
			sum.PIICollections++
		}
		sum.FindingsExtracted += len(cr.SampleDocuments)
	}
	sum.Severity = r.Severity
}
