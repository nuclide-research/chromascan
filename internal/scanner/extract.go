package scanner

import (
	"encoding/json"
	"fmt"
)

// sampleRecords posts to the /get endpoint and returns up to limit documents
// along with their metadata. Embeddings are never requested (performance).
func sampleRecords(c *httpClient, api apiVersion, tenant, database, colID string, limit int) (*chromaGetResponse, error) {
	var path string
	if api == apiV2 {
		path = fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/get",
			tenant, database, colID)
	} else {
		path = fmt.Sprintf("/api/v1/collections/%s/get", colID)
	}

	body := chromaGetRequest{
		Limit:   limit,
		Include: []string{"documents", "metadatas"},
	}

	var resp chromaGetResponse
	if _, err := c.post(path, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// parseJSON is a thin wrapper so other files can decode raw bytes uniformly.
func parseJSON(data []byte, out interface{}) error {
	return json.Unmarshal(data, out)
}
