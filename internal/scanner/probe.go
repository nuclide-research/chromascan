package scanner

import (
	"encoding/json"
	"fmt"
	"strings"
)

// apiVersion tracks which ChromaDB API generation is in use.
type apiVersion string

const (
	apiV2 apiVersion = "v2"
	apiV1 apiVersion = "v1"
)

// probeResult holds the output of the fingerprint phase.
type probeResult struct {
	version    string
	api        apiVersion
	authStatus string // "OPEN" or "AUTH_REQUIRED"
}

// probe fingerprints the target and determines auth status.
// It tries v2 first, falls back to v1.
func probe(c *httpClient) (*probeResult, error) {
	res := &probeResult{}

	// Try v2 heartbeat first.
	ver, api, err := detectVersion(c)
	if err != nil {
		return nil, fmt.Errorf("unreachable: %w", err)
	}
	res.version = ver
	res.api = api

	// Auth check: attempt to list collections. 401/403 = auth required.
	res.authStatus = checkAuth(c, api)

	return res, nil
}

// detectVersion tries v2 then v1 heartbeat + version endpoints.
func detectVersion(c *httpClient) (string, apiVersion, error) {
	// v2 heartbeat
	var hb heartbeatResponse
	if _, err := c.get("/api/v2/heartbeat", &hb); err == nil {
		ver := fetchVersion(c, "/api/v2/version")
		return ver, apiV2, nil
	}

	// v1 heartbeat
	if _, err := c.get("/api/v1/heartbeat", &hb); err == nil {
		ver := fetchVersion(c, "/api/v1/version")
		return ver, apiV1, nil
	}

	return "", "", fmt.Errorf("no heartbeat on v1 or v2")
}

// fetchVersion retrieves the raw version string from the version endpoint.
// ChromaDB returns a bare JSON string, e.g. "1.4.2".
func fetchVersion(c *httpClient, path string) string {
	body, code, err := c.getRaw(path)
	if err != nil || code != 200 {
		return "unknown"
	}
	var ver string
	if err := json.Unmarshal(body, &ver); err != nil {
		// Some deployments return plain text without quotes.
		return strings.TrimSpace(string(body))
	}
	return ver
}

// checkAuth lists collections and inspects the status code.
func checkAuth(c *httpClient, api apiVersion) string {
	var path string
	if api == apiV2 {
		path = "/api/v2/tenants/default_tenant/databases/default_database/collections"
	} else {
		path = "/api/v1/collections"
	}

	var dummy []rawCollection
	code, _ := c.get(path, &dummy)
	if code == 401 || code == 403 {
		return "AUTH_REQUIRED"
	}
	return "OPEN"
}
