// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Paging Details about paging
//
// swagger:model Paging
type Paging struct {

	// The URL of the request without any parameters.
	BaseURL string `json:"baseUrl,omitempty"`

	// The fields used in the original request
	Fields string `json:"fields,omitempty"`

	// The filter used in the original request
	Filter string `json:"filter,omitempty"`

	// The id of the last element that can be used to retrieve the next elements
	LastID int64 `json:"lastId,omitempty"`

	// The URL to retrieve the next page. Paging is enabled when more than 1000 objects would be returned by the query
	Next string `json:"next,omitempty"`
}

// Validate validates this paging
func (m *Paging) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Paging) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Paging) UnmarshalBinary(b []byte) error {
	var res Paging
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
