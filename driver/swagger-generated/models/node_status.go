// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NodeStatus Summary information about a Node
//
// swagger:model NodeStatus
type NodeStatus struct {

	// The state of GPFS on this node as reported by 'mmhealth node show'
	GpfsState string `json:"gpfsState,omitempty"`

	// The state of the node as reported by 'mmhealth node show'
	NodeState string `json:"nodeState,omitempty"`

	// The name of Operating System running on this node
	OsName string `json:"osName,omitempty"`

	// The GPFS version installed on this node
	ProductVersion string `json:"productVersion,omitempty"`
}

// Validate validates this node status
func (m *NodeStatus) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *NodeStatus) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *NodeStatus) UnmarshalBinary(b []byte) error {
	var res NodeStatus
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
