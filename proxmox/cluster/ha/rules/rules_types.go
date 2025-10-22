/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package rules

import (
	"github.com/bpg/terraform-provider-proxmox/proxmox/types"
)

// HARuleListResponseBody contains the body from a HA rule list response.
type HARuleListResponseBody struct {
	Data []*HARuleListResponseData `json:"data,omitempty"`
}

// HARuleListResponseData contains the data from a HA rule list response.
type HARuleListResponseData struct {
	Rule *string `json:"rule"`
}

// HARuleGetResponseBody contains the body from a HA rule get response.
type HARuleGetResponseBody struct {
	Data *HARuleGetResponseData `json:"data,omitempty"`
}

// HARuleDataBase contains data common to all HA rule API calls.
type HARuleDataBase struct {
	// Rule
	Rule *string `json:"rule,omitempty" url:"rule,omitempty"`
	// Rule type
	Type types.HARuleType `json:"type,omitempty" url:"type,omitempty"`
}

// HARuleGetResponseData contains data received from the HA rule API when requesting information about a single
// HA rule.
type HARuleGetResponseData struct {
	HARuleDataBase
}

// HARuleCreateRequestBody contains data received from the HA rule API when creating a new HA rule.
type HARuleCreateRequestBody struct {
	HARuleDataBase

	// A comma-separated list of HA resource fields. Each resource field contains a resource type and
	// a resource id, with a semicolon acting as a separator.
	Resources string `url:"resources"`
	// Identifier of this resource
	Affinity *string `url:"affinity,omitempty"`
	// A boolean (0/1) indicating whether the node affinity rule is strict or non-strict.
	Strict types.CustomBool `url:"type,int"`
	// Rule comment, if defined
	Comment *string `url:"comment,omitempty"`
	// Whether the HA rule is disabled.
	Disable types.CustomBool `url:"disable,int"`
	// A comma-separated list of node fields. Each node field contains a node name, and may
	// include a priority, with a semicolon acting as a separator.
	Nodes *string `url:"nodes,omitempty"`
}

// HARuleUpdateRequestBody contains data received from the HA rule API when updating an existing HA rule.
type HARuleUpdateRequestBody struct {
	HARuleDataBase

	// A comma-separated list of HA resource fields. Each resource field contains a resource type and
	// a resource id, with a semicolon acting as a separator.
	Resources *string `url:"resources,omitempty"`
	// Identifier of this resource
	Affinity *string `url:"affinity,omitempty"`
	// A boolean (0/1) indicating whether the node affinity rule is strict or non-strict.
	Strict types.CustomBool `url:"type,int"`
	// Rule comment, if defined
	Comment *string `url:"comment,omitempty"`
	// Whether the HA rule is disabled.
	Disable types.CustomBool `url:"disable,int"`
	// A comma-separated list of node fields. Each node field contains a node name, and may
	// include a priority, with a semicolon acting as a separator.
	Nodes *string `url:"nodes,omitempty"`
	// Prevent changes if current configuration file has a different digest.
	Digest *string `url:"digest,omitempty"`
	// Settings that must be deleted from the rule's configuration
	Delete []string `url:"delete,omitempty,comma"`
}
