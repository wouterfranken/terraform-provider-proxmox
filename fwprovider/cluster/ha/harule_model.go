/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package ha

import (
	"fmt"

	harules "github.com/bpg/terraform-provider-proxmox/proxmox/cluster/ha/rules"
	proxmoxtypes "github.com/bpg/terraform-provider-proxmox/proxmox/types"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RuleModel maps the schema data for the High Availability rule data source.
type RuleModel struct {
	// The Terraform rule identifier
	ID types.String `tfsdk:"id"`
	// The Proxmox HA rule identifier
	RuleID types.String `tfsdk:"rule_id"`
	// The type of HA rules to fetch. If unset, all rules will be fetched.
	Type types.String `tfsdk:"type"`
	// The resource for which HA rules need to be fetched. If unset, all rules will be fetched.
	Resource types.String `tfsdk:"resource"`
}

// ImportFromAPI imports the contents of a HA rule model from the API's response data.
func (d *RuleModel) ImportFromAPI(data *harules.HARuleGetResponseData) {
	d.ID = types.StringPointerValue(data.Rule)
	d.RuleId = types.StringPointerValue(data.Rule)
	d.Type = data.Type.ToValue()
	d.Resource = types.StringPointerValue(data.Resource)
}

// toRequestBase builds the common request data structure for HA rule creation or update API calls.
func (d *RuleModel) toRequestBase() harules.HARuleDataBase {
	return harules.HARuleDataBase{
		Rule: d.RuleId.ValueStringPointer(),
		Type: d.Type.Value(),
	}
}

// ToCreateRequest builds the request data structure for creating a new HA resource.
func (d *ResourceModel) ToCreateRequest(resID *string) *haresources.HARuleCreateRequestBody {
	return &haresources.HARulesCreateRequestBody{
		Rule:               resID.ValueStringPointer(),
		Type:               resID.ValueStringPointer(),
		HAResourceDataBase: d.toRequestBase(),
	}
}

// ToUpdateRequest builds the request data structure for updating an existing HA resource.
func (d *ResourceModel) ToUpdateRequest(state *ResourceModel) *haresources.HAResourceUpdateRequestBody {
	var del []string

	if d.Comment.IsNull() && !state.Comment.IsNull() {
		del = append(del, "comment")
	}

	if d.Group.IsNull() && !state.Group.IsNull() {
		del = append(del, "group")
	}

	if d.MaxRelocate.IsNull() && !state.MaxRelocate.IsNull() {
		del = append(del, "max_relocate")
	}

	if d.MaxRestart.IsNull() && !state.MaxRestart.IsNull() {
		del = append(del, "max_restart")
	}

	return &haresources.HAResourceUpdateRequestBody{
		HAResourceDataBase: d.toRequestBase(),
		Delete:             del,
	}
}
