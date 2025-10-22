/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package ha

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bpg/terraform-provider-proxmox/fwprovider/attribute"
	"github.com/bpg/terraform-provider-proxmox/fwprovider/config"
	harules "github.com/bpg/terraform-provider-proxmox/proxmox/cluster/ha/rules"
	proxmoxtypes "github.com/bpg/terraform-provider-proxmox/proxmox/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// haRuleResource contains the rule's internal data.
type haRuleResource struct {
	// The HA rules API client
	client *harules.Client
}

// Ensure the resource implements the expected interfaces.
var (
	_ resource.Resource                = &haRuleResource{}
	_ resource.ResourceWithConfigure   = &haRuleResource{}
	_ resource.ResourceWithImportState = &haRuleResource{}
)

// NewHARuleResource returns a new resource for managing High Availability resources.
func NewHARuleResource() resource.Resource {
	return &haRuleResource{}
}

// Metadata defines the name of the resource.
func (r *haRuleResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_harule"
}

// Schema defines the schema for the resource.
func (r *haRuleResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Manages Proxmox HA rules.",
		Attributes: map[string]schema.Attribute{
			"id": attribute.ResourceID(),
			"rule_id": schema.StringAttribute{
				Description: "The Proxmox HA rule identifier",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^ha-rule-[a-z0-9]{8}-[a-z0-9]{4}$`),
							"must be a valid Proxmox rule id",
					),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of the rule.",
				Required:    true,
				Validators: []validator.String{
					ruleTypeValidator(),
				},
			},
			"resources": schema.ListAttribute{
				Description: "The member nodes for this group. They are provided as a map, where the keys are the node " +
					"names and the values represent their priority: integers for known priorities or `null` for unset " +
					"priorities.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					mapvalidator.SizeAtLeast(1),
					mapvalidator.ValueStringsAre(resourceIdValidator()),
				},
			},
			"comment": schema.StringAttribute{
				Description: "The comment associated with this resource.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.UTF8LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^\S|^$`), "must not start with whitespace"),
					stringvalidator.RegexMatches(regexp.MustCompile(`\S$|^$`), "must not end with whitespace"),
				},
			},
			"affinity": schema.StringAttribute{
				Description: "The affinity of the rule.",
				Required:    true,
				Validators: []validator.String{
					ruleAffinityValidator(),
				},
			},
			"nodes": schema.MapAttribute{
				Description: "The member nodes for this group. They are provided as a map, where the keys are the node " +
					"names and the values represent their priority: integers for known priorities or `null` for unset " +
					"priorities.",
				Required:    true,
				ElementType: types.Int64Type,
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
					mapvalidator.KeysAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?$`),
							"must be a valid Proxmox node name",
						),
					),
					mapvalidator.ValueInt64sAre(int64validator.Between(0, 1000)),
				},
			},
			"strict": schema.BoolAttribute{
				Description: "A flag that indicates that failing back to a higher priority node is disabled for this HA " +
					"group. Defaults to `false`.",
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			"disable": schema.BoolAttribute{
				Description: "A flag that indicates that failing back to a higher priority node is disabled for this HA " +
					"group. Defaults to `false`.",
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider-configured client to the resource.
func (r *haResourceResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected config.Resource, got: %T", req.ProviderData),
		)

		return
	}

	r.client = cfg.Client.Cluster().HA().Resources()
}

// Create creates a new HA resource.
func (r *haResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resID, err := proxmoxtypes.ParseHAResourceID(data.ResourceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected error parsing Proxmox HA resource identifier",
			fmt.Sprintf("Couldn't parse the Terraform resource ID into a valid HA resource identifier: %s", err),
		)

		return
	}

	createRequest := data.ToCreateRequest(resID)

	err = r.client.Create(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not create HA resource '%v'.", resID),
			err.Error(),
		)

		return
	}

	data.ID = types.StringValue(resID.String())

	r.readBack(ctx, &data, &resp.Diagnostics, &resp.State)
}

// Update updates an existing HA resource.
func (r *haResourceResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var data, state ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resID, err := proxmoxtypes.ParseHAResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected error parsing Proxmox HA resource identifier",
			fmt.Sprintf("Couldn't parse the Terraform resource ID into a valid HA resource identifier: %s", err),
		)

		return
	}

	updateRequest := data.ToUpdateRequest(&state)

	err = r.client.Update(ctx, resID, updateRequest)
	if err == nil {
		r.readBack(ctx, &data, &resp.Diagnostics, &resp.State)
	} else {
		resp.Diagnostics.AddError(
			"Error updating HA resource",
			fmt.Sprintf("Could not update HA resource '%s', unexpected error: %s",
				state.Group.ValueString(), err.Error()),
		)
	}
}

// Delete deletes an existing HA resource.
func (r *haResourceResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resID, err := proxmoxtypes.ParseHAResourceID(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected error parsing Proxmox HA resource identifier",
			fmt.Sprintf("Couldn't parse the Terraform resource ID into a valid HA resource identifier: %s", err),
		)

		return
	}

	err = r.client.Delete(ctx, resID)
	if err != nil {
		if strings.Contains(err.Error(), "no such resource") {
			resp.Diagnostics.AddWarning(
				"HA resource does not exist",
				fmt.Sprintf(
					"Could not delete HA resource '%v', it does not exist or has been deleted outside of Terraform.",
					resID,
				),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting HA resource",
				fmt.Sprintf("Could not delete HA resource '%v', unexpected error: %s",
					resID, err.Error()),
			)
		}
	}
}

// Read reads the HA resource.
func (r *haResourceResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.read(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if !resp.Diagnostics.HasError() {
		if found {
			resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
		} else {
			resp.State.RemoveResource(ctx)
		}
	}
}

// ImportState imports a HA resource from the Proxmox cluster.
func (r *haResourceResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	reqID := req.ID
	data := ResourceModel{
		ID:         types.StringValue(reqID),
		ResourceID: types.StringValue(reqID),
	}
	r.readBack(ctx, &data, &resp.Diagnostics, &resp.State)
}

// read reads information about a HA resource from the cluster. The Terraform resource identifier must have been set
// in the model before this function is called.
func (r *haResourceResource) read(ctx context.Context, data *ResourceModel) (bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	resID, err := proxmoxtypes.ParseHAResourceID(data.ID.ValueString())
	if err != nil {
		diags.AddError(
			"Unexpected error parsing Proxmox HA resource identifier",
			fmt.Sprintf("Couldn't parse the Terraform resource ID into a valid HA resource identifier: %s", err),
		)

		return false, diags
	}

	res, err := r.client.Get(ctx, resID)
	if err != nil {
		if !strings.Contains(err.Error(), "no such resource") {
			diags.AddError("Could not read HA resource", err.Error())
		}

		return false, diags
	}

	data.ImportFromAPI(res)

	return true, nil
}

// readBack reads information about a created or modified HA resource from the cluster then updates the response
// state accordingly. It is assumed that the `state`'s identifier is set.
func (r *haResourceResource) readBack(
	ctx context.Context,
	data *ResourceModel,
	respDiags *diag.Diagnostics,
	respState *tfsdk.State,
) {
	found, diags := r.read(ctx, data)

	respDiags.Append(diags...)

	if !found {
		respDiags.AddError(
			"HA resource not found after update",
			"Failed to find the resource when trying to read back the updated HA resource's data.",
		)
	}

	if !respDiags.HasError() {
		respDiags.Append(respState.Set(ctx, *data)...)
	}
}
