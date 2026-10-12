package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// usersPath is the SCIM2 Users endpoint exposed by the Unity Catalog control
// plane.
const usersPath = "/api/1.0/unity-control/scim2/Users"

// Email represents a SCIM2 user email.
type Email struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary"`
}

// User represents a SCIM2 user resource. Only the attributes managed by the
// Terraform provider are modelled here; any additional attributes returned by
// the server are ignored during decoding.
type User struct {
	// ID is assigned by the server. It is omitted on create (POST) but must be
	// present in the body of an update (PUT), even though it is not user
	// configurable.
	ID          string   `json:"id,omitempty"`
	Schemas     []string `json:"schemas"`
	DisplayName string   `json:"displayName,omitempty"`
	Emails      []Email  `json:"emails,omitempty"`
	Active      bool     `json:"active"`
}

// usersURL returns the full URL of the SCIM2 Users collection.
func (c *Client) usersURL() string {
	return c.host + usersPath
}

// userURL returns the full URL of a single SCIM2 user resource.
func (c *Client) userURL(id string) string {
	return c.usersURL() + "/" + url.PathEscape(id)
}

// CreateUser creates a user via POST /scim2/Users and returns the created user.
func (c *Client) CreateUser(ctx context.Context, u *User) (*User, error) {
	resp, err := c.do(ctx, http.MethodPost, c.usersURL(), u)
	if err != nil {
		return nil, err
	}
	body, err := readResponse(resp, http.StatusCreated, "create user")
	if err != nil {
		return nil, err
	}

	var created User
	if err := decodeOrEmpty(body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetUser fetches a user via GET /scim2/Users/{id}. It returns (nil, nil) when
// the user does not exist.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	resp, err := c.do(ctx, http.MethodGet, c.userURL(id), nil)
	if err != nil {
		return nil, err
	}
	body, err := readResponse(resp, http.StatusOK, "get user")
	if err != nil {
		return nil, err
	}

	var u User
	if err := decodeOrEmpty(body, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUser replaces a user via PUT /scim2/Users/{id}. The ID field of u must
// be set; it is included in the request body as required by the SCIM2 server,
// even though the provider never allows the user to change it. It returns
// (nil, nil) when the user no longer exists.
func (c *Client) UpdateUser(ctx context.Context, u *User) (*User, error) {
	if u.ID == "" {
		return nil, fmt.Errorf("update user requires a non-empty id")
	}
	resp, err := c.do(ctx, http.MethodPut, c.userURL(u.ID), u)
	if err != nil {
		return nil, err
	}
	body, err := readResponse(resp, http.StatusOK, "update user")
	if err != nil {
		return nil, err
	}

	// Some SCIM2 implementations respond with 204 No Content; fall back to the
	// input user so the caller always has a value to persist.
	var updated User
	if err := decodeOrEmpty(body, &updated); err != nil {
		return nil, err
	}
	if updated.ID == "" {
		updated = *u
	}
	return &updated, nil
}

// DeleteUser deletes a user via DELETE /scim2/Users/{id}. A 404 response is
// treated as success (the resource is already gone).
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	resp, err := c.do(ctx, http.MethodDelete, c.userURL(id), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		readBody(resp)
		return nil
	}
	_, err = readResponse(resp, http.StatusOK, "delete user")
	return err
}
