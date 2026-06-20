package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type httpClient struct {
	baseURL string
	http    *http.Client
	verbose bool
}

func newClient(baseURL string, timeoutSec int, verbose bool) *httpClient {
	return &httpClient{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		verbose: verbose,
	}
}

// get fetches a path and decodes JSON into out. Returns (statusCode, error).
func (c *httpClient) get(path string, out interface{}) (int, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}
	if c.verbose {
		fmt.Printf("  GET %s -> %d (%d bytes)\n", path, resp.StatusCode, len(body))
	}
	if resp.StatusCode != 200 {
		return resp.StatusCode, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp.StatusCode, json.Unmarshal(body, out)
}

// getRaw fetches a path and returns raw bytes plus status code.
func (c *httpClient) getRaw(path string) ([]byte, int, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// post sends JSON body to path and decodes the response into out.
func (c *httpClient) post(path string, payload interface{}, out interface{}) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}
	if c.verbose {
		fmt.Printf("  POST %s -> %d (%d bytes)\n", path, resp.StatusCode, len(body))
	}
	if resp.StatusCode >= 400 {
		return resp.StatusCode, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	if out != nil {
		return resp.StatusCode, json.Unmarshal(body, out)
	}
	return resp.StatusCode, nil
}
