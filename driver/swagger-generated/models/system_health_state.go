// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// SystemHealthState Summary information about a System Health event
//
// swagger:model SystemHealthState
type SystemHealthState struct {

	// The time since when this state is active
	ActiveSince string `json:"activeSince,omitempty"`

	// The component the state belongs to
	Component string `json:"component,omitempty"`

	// The name of the entity the this state belongs to
	EntityName string `json:"entityName,omitempty"`

	// The type of the entity this state belongs to
	EntityType string `json:"entityType,omitempty"`

	// The internal unique id of the state
	Oid int32 `json:"oid,omitempty"`

	// The name of the parent of the entity this state belongs to
	ParentName string `json:"parentName,omitempty"`

	// A list of events which led to this state (only set for non-healthy states)
	Reasons []string `json:"reasons"`

	// The name of the node the state belongs to
	ReportingNode string `json:"reportingNode,omitempty"`

	// The state of the component and entity on the specified reporting node
	State string `json:"state,omitempty"`
}

// Validate validates this system health state
func (m *SystemHealthState) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *SystemHealthState) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SystemHealthState) UnmarshalBinary(b []byte) error {
	var res SystemHealthState
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
