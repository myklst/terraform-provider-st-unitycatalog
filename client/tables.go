// Copyright IBM Corp. 2026

package client

import (
	"context"
	"net/http"
	"net/url"
)

// tablesPath is the Unity Catalog tables list endpoint.
const tablesPath = "/api/2.1/unity-catalog/tables"

// TableInfo represents a single table returned by the Unity Catalog tables
// list endpoint. Only commonly used fields are modelled.
type TableInfo struct {
	CatalogName string `json:"catalog_name"`
	SchemaName  string `json:"schema_name"`
	Name        string `json:"name"`
	TableType   string `json:"table_type"`
	Comment     string `json:"comment,omitempty"`
}

// ListTablesResponse is the body returned by GET /tables.
type ListTablesResponse struct {
	Tables        []TableInfo `json:"tables"`
	NextPageToken string      `json:"next_page_token,omitempty"`
}

// ListTables lists tables in a catalog (optionally filtered by schema) via
// GET /api/2.1/unity-catalog/tables?catalog_name=...&schema_name=...
func (c *Client) ListTables(ctx context.Context, catalogName, schemaName string) (*ListTablesResponse, error) {
	params := url.Values{}
	params.Set("catalog_name", catalogName)
	if schemaName != "" {
		params.Set("schema_name", schemaName)
	}

	fullURL := c.host + tablesPath + "?" + params.Encode()

	resp, err := c.doWithContentType(ctx, http.MethodGet, fullURL, nil, "application/json")
	if err != nil {
		return nil, err
	}
	body, err := readResponse(resp, http.StatusOK, "list tables")
	if err != nil {
		return nil, err
	}

	var result ListTablesResponse
	if err := decodeOrEmpty(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
