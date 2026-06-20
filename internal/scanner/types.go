package scanner

import "time"

// Config holds all runtime configuration.
type Config struct {
	Target        string
	Tenant        string
	Database      string
	TimeoutSec    int
	DoWriteCanary bool
	ProbeOnly     bool
	OutputFile    string
	Verbose       bool
}

// Raw ChromaDB API response shapes

type heartbeatResponse struct {
	NanosecondHeartbeat int64 `json:"nanosecond heartbeat"`
}

// rawCollection is shared between v1 and v2 collection list responses.
type rawCollection struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Dimension         *int                   `json:"dimension"`
	Metadata          map[string]interface{} `json:"metadata"`
	ConfigurationJSON map[string]interface{} `json:"configuration_json"`
}

// chromaGetRequest is the body for POST .../collections/{id}/get
type chromaGetRequest struct {
	Limit   int      `json:"limit"`
	Include []string `json:"include"`
}

// chromaGetResponse is the body returned by POST .../get
type chromaGetResponse struct {
	IDs       []string                 `json:"ids"`
	Documents []string                 `json:"documents"`
	Metadatas []map[string]interface{} `json:"metadatas"`
}

// chromaAddRequest is the body for POST .../add (write canary)
type chromaAddRequest struct {
	IDs        []string                 `json:"ids"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas"`
	Embeddings [][]float64              `json:"embeddings"`
}

// chromaDeleteRequest is the body for POST .../delete (canary cleanup)
type chromaDeleteRequest struct {
	IDs []string `json:"ids"`
}

// Output types

// CollectionResult holds enumeration output for one ChromaDB collection.
type CollectionResult struct {
	Name            string                   `json:"name"`
	ID              string                   `json:"id"`
	RecordCount     int64                    `json:"record_count"`
	Dimension       int                      `json:"dimension,omitempty"`
	MetadataKeys    []string                 `json:"metadata_keys,omitempty"`
	PIISignals      []string                 `json:"pii_signals,omitempty"`
	SampleDocuments []string                 `json:"sample_documents,omitempty"`
	SampleMetadatas []map[string]interface{} `json:"sample_metadatas,omitempty"`
	Score           float64                  `json:"score"`
	Severity        string                   `json:"severity"`
}

// CanaryResult records the write canary outcome.
type CanaryResult struct {
	Attempted bool   `json:"attempted"`
	Success   bool   `json:"success"`
	CanaryID  string `json:"canary_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ScanResult is the top-level output object.
type ScanResult struct {
	Target      string              `json:"target"`
	ScanTime    time.Time           `json:"scan_time"`
	Version     string              `json:"version"`
	APIVersion  string              `json:"api_version"`
	AuthStatus  string              `json:"auth_status"`
	Collections []CollectionResult  `json:"collections,omitempty"`
	WriteTest   string              `json:"write_test,omitempty"`
	Canary      *CanaryResult       `json:"canary,omitempty"`
	Score       float64             `json:"score"`
	Severity    string              `json:"severity"`
	Summary     ScanSummary         `json:"summary"`
}

// ScanSummary is the rolled-up finding counts.
type ScanSummary struct {
	TotalCollections      int    `json:"total_collections"`
	TotalRecordsEstimated int64  `json:"total_records_estimated"`
	PIICollections        int    `json:"pii_collections"`
	FindingsExtracted     int    `json:"findings_extracted"`
	Severity              string `json:"severity"`
	AuthStatus            string `json:"auth_status"`
}
