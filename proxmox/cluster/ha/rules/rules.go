/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package resources

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/bpg/terraform-provider-proxmox/proxmox/api"
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

type haRuleListQueryParams struct {
	ResType *types.HARuleType `url:"type"`
	ResResource *string `url:"resource"`
}

// List retrieves the list of HA rules. If the `resType` and `resResource` arguments are `nil`, all rules will be returned;
// otherwise rules will be filtered by the specified type (either `node-affinity` or `resource-affinity`)
// and/or by the specified resource.
func (c *Client) List(ctx context.Context, resType *types.HARuleType, resResource *string) ([]*HARuleListResponseData, error) {
	options := &haRuleListQueryParams{resType, resResource}
	resBody := &HARuleListResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath(""), options, resBody)
	if err != nil {
		return nil, fmt.Errorf("error listing HA rules: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	sort.Slice(resBody.Data, func(i, j int) bool {
		return resBody.Data[i].Rule < resBody.Data[j].Rule
	})

	return resBody.Data, nil
}

// Get retrieves the configuration of a single HA rule.
func (c *Client) Get(ctx context.Context, rule *string) (*HARuleGetResponseData, error) {
	resBody := &HARuleGetResponseBody{}

	err := c.DoRequest(ctx, http.MethodGet, c.ExpandPath(url.PathEscape(rule)), nil, resBody)
	if err != nil {
		return nil, fmt.Errorf("error reading HA resource: %w", err)
	}

	if resBody.Data == nil {
		return nil, api.ErrNoDataObjectInResponse
	}

	return resBody.Data, nil
}

// Create creates a new HA rule.
func (c *Client) Create(ctx context.Context, data *HARuleCreateRequestBody) error {
	err := c.DoRequest(ctx, http.MethodPost, c.ExpandPath(""), data, nil)
	if err != nil {
		return fmt.Errorf("error creating HA resource: %w", err)
	}

	return nil
}

// Update updates an existing HA rule.
func (c *Client) Update(ctx context.Context, rule *string, data *HARuleUpdateRequestBody) error {
	err := c.DoRequest(ctx, http.MethodPut, c.ExpandPath(url.PathEscape(rule)), data, nil)
	if err != nil {
		return fmt.Errorf("error updating HA resource %v: %w", id, err)
	}

	return nil
}

// Delete deletes a HA rule.
func (c *Client) Delete(ctx context.Context, rule *string) error {
	err := c.DoRequest(ctx, http.MethodDelete, c.ExpandPath(url.PathEscape(rule)), nil, nil)
	if err != nil {
		return fmt.Errorf("error deleting HA resource %v: %w", id, err)
	}

	return nil
}
