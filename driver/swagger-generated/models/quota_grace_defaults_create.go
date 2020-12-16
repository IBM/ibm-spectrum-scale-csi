// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// QuotaGraceDefaultsCreate quota grace defaults create
//
// swagger:model QuotaGraceDefaultsCreate
type QuotaGraceDefaultsCreate struct {

	// block grace period
	BlockGracePeriod string `json:"blockGracePeriod,omitempty"`

	// files grace period
	FilesGracePeriod string `json:"filesGracePeriod,omitempty"`

	// grace
	Grace string `json:"grace,omitempty"`
}

// Validate validates this quota grace defaults create
func (m *QuotaGraceDefaultsCreate) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *QuotaGraceDefaultsCreate) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *QuotaGraceDefaultsCreate) UnmarshalBinary(b []byte) error {
	var res QuotaGraceDefaultsCreate
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
