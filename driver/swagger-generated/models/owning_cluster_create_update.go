// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// OwningClusterCreateUpdate Information for registering a cluster that owns filesystems to be mounted remotely
//
// swagger:model OwningClusterCreateUpdate
type OwningClusterCreateUpdate struct {

	// The contact nodes of the owning cluster used for remote mounting
	ContactNodes []string `json:"contactNodes"`

	// The public RSA key of the owning cluster
	// Required: true
	Key []string `json:"key"`

	// The owning cluster of the remote filesystem
	// Required: true
	OwningCluster *string `json:"owningCluster"`
}

// Validate validates this owning cluster create update
func (m *OwningClusterCreateUpdate) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateKey(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOwningCluster(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *OwningClusterCreateUpdate) validateKey(formats strfmt.Registry) error {

	if err := validate.Required("key", "body", m.Key); err != nil {
		return err
	}

	return nil
}

func (m *OwningClusterCreateUpdate) validateOwningCluster(formats strfmt.Registry) error {

	if err := validate.Required("owningCluster", "body", m.OwningCluster); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *OwningClusterCreateUpdate) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *OwningClusterCreateUpdate) UnmarshalBinary(b []byte) error {
	var res OwningClusterCreateUpdate
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
