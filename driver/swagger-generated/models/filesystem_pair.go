// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// FilesystemPair The two filesystem names that are having a remote mount relationship
//
// swagger:model FilesystemPair
type FilesystemPair struct {

	// The filesystem on the owning cluster.
	OwningClusterFilesystem string `json:"owningClusterFilesystem,omitempty"`

	// The filesystem on the remote cluster.
	RemoteClusterFilesystem string `json:"remoteClusterFilesystem,omitempty"`
}

// Validate validates this filesystem pair
func (m *FilesystemPair) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *FilesystemPair) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *FilesystemPair) UnmarshalBinary(b []byte) error {
	var res FilesystemPair
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
