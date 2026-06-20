package scanner

import "fmt"

// listCollections fetches all collections, falling back from v2 to v1.
func listCollections(c *httpClient, api apiVersion, tenant, database string) ([]rawCollection, error) {
	if api == apiV2 {
		path := fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections", tenant, database)
		var cols []rawCollection
		code, err := c.get(path, &cols)
		if err == nil {
			return cols, nil
		}
		if code == 404 {
			// Fall through to v1.
			api = apiV1
		} else {
			return nil, err
		}
	}

	// v1 fallback
	var cols []rawCollection
	if _, err := c.get("/api/v1/collections", &cols); err != nil {
		return nil, err
	}
	return cols, nil
}

// fetchCount gets the integer record count for a collection by ID.
func fetchCount(c *httpClient, api apiVersion, tenant, database, colID string) int64 {
	var path string
	if api == apiV2 {
		path = fmt.Sprintf("/api/v2/tenants/%s/databases/%s/collections/%s/count",
			tenant, database, colID)
	} else {
		path = fmt.Sprintf("/api/v1/collections/%s/count", colID)
	}

	// The count endpoint returns a bare integer, not a JSON object.
	body, code, err := c.getRaw(path)
	if err != nil || code != 200 {
		return 0
	}
	var count int64
	if err := parseJSON(body, &count); err != nil {
		return 0
	}
	return count
}
