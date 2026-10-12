// Copyright IBM Corp. 2026

package client

import (
	"context"
	"net/http"
)

// permissionsPath is the Unity Catalog permissions endpoint.
const permissionsPath = "/api/2.1/unity-catalog/permissions"

// PrivilegeChange represents a single privilege change for a principal. Add
// and Remove are optional; include only the relevant field for the operation.
type PrivilegeChange struct {
	Principal string   `json:"principal"`
	Add       []string `json:"add,omitempty"`
	Remove    []string `json:"remove,omitempty"`
}

// UpdatePermissionsRequest is the body for PUT /permissions/{type}/{name}.
type UpdatePermissionsRequest struct {
	Changes []PrivilegeChange `json:"changes"`
}

// PrivilegeAssignment represents a principal's privileges on a securable, as
// returned by GET /permissions/{type}/{name}.
type PrivilegeAssignment struct {
	Principal  string   `json:"principal"`
	Privileges []string `json:"privileges"`
}

// GetPermissionsResponse is the body returned by GET /permissions/{type}/{name}.
type GetPermissionsResponse struct {
	PrivilegeAssignments []PrivilegeAssignment `json:"privilege_assignments"`
}

// permissionsURL returns the full URL for a securable's permissions.
func (c *Client) permissionsURL(securableType, fullName string) string {
	return c.host + permissionsPath + "/" + securableType + "/" + fullName
}

// UpdatePermissions applies privilege changes via PUT
// /permissions/{securableType}/{fullName}. It returns an error on failure.
func (c *Client) UpdatePermissions(ctx context.Context, securableType, fullName string, req *UpdatePermissionsRequest) error {
	resp, err := c.doWithContentType(ctx, http.MethodPatch, c.permissionsURL(securableType, fullName), req, "application/json")
	if err != nil {
		return err
	}
	_, err = readResponse(resp, http.StatusOK, "update permissions")
	return err
}

// GetPermissions fetches all privilege assignments for a securable via GET
// /permissions/{securableType}/{fullName}. It returns (nil, nil) when the
// securable does not exist.
func (c *Client) GetPermissions(ctx context.Context, securableType, fullName string) (*GetPermissionsResponse, error) {
	resp, err := c.doWithContentType(ctx, http.MethodGet, c.permissionsURL(securableType, fullName), nil, "application/json")
	if err != nil {
		return nil, err
	}
	body, err := readResponse(resp, http.StatusOK, "get permissions")
	if err != nil {
		return nil, err
	}

	var result GetPermissionsResponse
	if err := decodeOrEmpty(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
