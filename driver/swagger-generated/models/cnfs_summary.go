// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// CnfsSummary Summary information about cNFS
//
// swagger:model CnfsSummary
type CnfsSummary struct {

	// The IP of the Ganesha NFS server
	CnfsGanesha string `json:"cnfsGanesha,omitempty"`

	// Specifies if cnfs monitoring is enabled
	CnfsMonitorEnabled string `json:"cnfsMonitorEnabled,omitempty"`

	// The tcp port used by the mount daemon
	CnfsMountdPort string `json:"cnfsMountdPort,omitempty"`

	// The number of NFS kernel processes
	CnfsNFSDprocs string `json:"cnfsNFSDprocs,omitempty"`

	// Specified if node will reboot if monitoring detects an unrecoverable problem
	CnfsReboot string `json:"cnfsReboot,omitempty"`

	// The shared path where the cnfs stores internal state
	CnfsSharedRoot string `json:"cnfsSharedRoot,omitempty"`
}

// Validate validates this cnfs summary
func (m *CnfsSummary) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *CnfsSummary) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CnfsSummary) UnmarshalBinary(b []byte) error {
	var res CnfsSummary
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
