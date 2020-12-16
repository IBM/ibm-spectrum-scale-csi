// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NodeClassInlineResponse node class inline response
//
// swagger:model NodeClassInlineResponse
type NodeClassInlineResponse struct {

	// nodeclasses
	NodeClasses []*NodeClass `json:"nodeClasses"`

	// paging
	Paging *Paging `json:"paging,omitempty"`

	// The status of the request
	Status *RequestStatus `json:"status,omitempty"`
}

// Validate validates this node class inline response
func (m *NodeClassInlineResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateNodeClasses(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePaging(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *NodeClassInlineResponse) validateNodeClasses(formats strfmt.Registry) error {

	if swag.IsZero(m.NodeClasses) { // not required
		return nil
	}

	for i := 0; i < len(m.NodeClasses); i++ {
		if swag.IsZero(m.NodeClasses[i]) { // not required
			continue
		}

		if m.NodeClasses[i] != nil {
			if err := m.NodeClasses[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("nodeClasses" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *NodeClassInlineResponse) validatePaging(formats strfmt.Registry) error {

	if swag.IsZero(m.Paging) { // not required
		return nil
	}

	if m.Paging != nil {
		if err := m.Paging.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("paging")
			}
			return err
		}
	}

	return nil
}

func (m *NodeClassInlineResponse) validateStatus(formats strfmt.Registry) error {

	if swag.IsZero(m.Status) { // not required
		return nil
	}

	if m.Status != nil {
		if err := m.Status.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("status")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *NodeClassInlineResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *NodeClassInlineResponse) UnmarshalBinary(b []byte) error {
	var res NodeClassInlineResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
