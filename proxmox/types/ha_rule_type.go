/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package types

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/google/go-querystring/query"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HARuleType represents the type of HA rule.
type HARuleType int

// Ensure various interfaces are supported by the HA rule type type.
// NOTE: to my knowledge, this "global" here is required for the static type checks to work.
var (
	//nolint:gochecknoglobals
	_haRuleTypeValue HARuleType
	_                fmt.Stringer     = &_haRuleTypeValue
	_                json.Marshaler   = &_haRuleTypeValue
	_                json.Unmarshaler = &_haRuleTypeValue
	_                query.Encoder    = &_haRuleTypeValue
)

const (
	// HARuleTypeNodeAffinity indicates that a HA rule is a node affinity rule.
	HARuleTypeNodeAffinity HARuleType = 0
	// HARuleTypeResourceAffinity indicates that a HA rule is a resource affinity rule.
	HARuleTypeResourceAffinity HARuleType = 1
)

// ParseHARuleType converts the string representation of a HA rule type into the corresponding
// enum value. An error is returned if the input string does not match any known type.
func ParseHARuleType(input string) (HARuleType, error) {
	switch input {
	case "node-affinity":
		return HARuleTypeNodeAffinity, nil
	case "resource-affinity":
		return HARuleTypeResourceAffinity, nil
	default:
		return _haRuleTypeValue, fmt.Errorf("illegal HA rule type '%s'", input)
	}
}

// String converts a HARuleType value into a string.
func (t HARuleType) String() string {
	switch t {
	case HARuleTypeNodeAffinity:
		return "node-affinity"
	case HARuleTypeResourceAffinity:
		return "resource-affinity"
	default:
		panic(fmt.Sprintf("unknown HA rule type value: %d", t))
	}
}

// MarshalJSON marshals a HA rule type into JSON value.
func (t HARuleType) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(t.String())
	if err != nil {
		return nil, fmt.Errorf("cannot marshal HA rule type: %w", err)
	}

	return bytes, nil
}

// UnmarshalJSON unmarshals a Proxmox HA rule type.
func (t *HARuleType) UnmarshalJSON(b []byte) error {
	var rtString string

	err := json.Unmarshal(b, &rtString)
	if err != nil {
		return fmt.Errorf("cannot unmarshal HA rule type: %w", err)
	}

	resType, err := ParseHARuleType(rtString)
	if err == nil {
		*t = resType
	}

	return err
}

// EncodeValues encodes a HA rule type field into an URL-encoded set of values.
func (t HARuleType) EncodeValues(key string, v *url.Values) error {
	v.Add(key, t.String())
	return nil
}

// ToValue converts a HA rule type into a Terraform value.
func (t HARuleType) ToValue() types.String {
	return types.StringValue(t.String())
}
