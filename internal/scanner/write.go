package scanner

import (
	"fmt"
	"time"
)

// injectCanary writes a single canary document to the first available collection,
// then immediately deletes it to leave no persistent trace.
// Requires DoWriteCanary + authorize-write flags (enforced in main).
func injectCanary(c *httpClient, api apiVersion, tenant, database string, cols []rawCollection) *CanaryResult {
	res := &CanaryResult{Attempted: true}

	if len(cols) == 0 {
		res.Error = "no collections available for canary write"
		return res
	}

	col := cols[0]
	canaryID := fmt.Sprintf("nuclide-canary-%d", time.Now().Unix())

	addPath := collectionPath(api, tenant, database, col.ID, "add")
	delPath := collectionPath(api, tenant, database, col.ID, "delete")

	// Build zero vector matching collection dimension (or dim=3 if null).
	// ChromaDB v1.0.0 requires embeddings to be [][]float64, not null.
	dim := 3
	if col.Dimension != nil && *col.Dimension > 0 {
		dim = *col.Dimension
	}
	vec := make([]float64, dim)
	addBody := chromaAddRequest{
		IDs:       []string{canaryID},
		Documents: []string{"nuclide-security-canary"},
		Metadatas: []map[string]interface{}{
			{"source": "nuclide-recon"},
		},
		Embeddings: [][]float64{vec},
	}

	if _, err := c.post(addPath, addBody, nil); err != nil {
		res.Error = fmt.Sprintf("add failed: %v", err)
		return res
	}

	// Write confirmed -- clean up immediately.
	res.Success = true
	res.CanaryID = canaryID

	delBody := chromaDeleteRequest{IDs: []string{canaryID}}
	c.post(delPath, delBody, nil) // best-effort delete; ignore error

	return res
}

// collectionPath builds the API path for a collection sub-endpoint.
func collectionPath(api apiVersion, tenant, database, colID, endpoint string) string {
	if api == apiV2 {
		return fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/%s",
			tenant, database, colID, endpoint)
	}
	return fmt.Sprintf("/api/v1/collections/%s/%s", colID, endpoint)
}
