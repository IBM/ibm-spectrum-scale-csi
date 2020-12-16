// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// VDisk Detailed information about a vdisk configuration
//
// swagger:model VDisk
type VDisk struct {

	// The granularity of the checksum
	ChecksumGranularity int64 `json:"checksumGranularity,omitempty"`

	// The name of the declustered array
	DeclusteredArray string `json:"declusteredArray,omitempty"`

	// The healthState status of the vdisk
	Health string `json:"health,omitempty"`

	// The name of the vdisk
	Name string `json:"name,omitempty"`

	// The code of the raid
	RaidCode string `json:"raidCode,omitempty"`

	// The name of the recovery group
	RecoveryGroup string `json:"recoveryGroup,omitempty"`

	// The remarks of the vdisk
	Remarks string `json:"remarks,omitempty"`

	// The size of the vdisk
	Size int64 `json:"size,omitempty"`

	// The state of the vdisk
	State string `json:"state,omitempty"`

	// The size of the track
	TrackSize int64 `json:"trackSize,omitempty"`
}

// Validate validates this v disk
func (m *VDisk) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VDisk) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VDisk) UnmarshalBinary(b []byte) error {
	var res VDisk
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
