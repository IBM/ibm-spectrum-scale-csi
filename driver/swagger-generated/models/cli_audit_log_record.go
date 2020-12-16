// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// CliAuditLogRecord Summary information about a CLI audit record
//
// swagger:model CliAuditLogRecord
type CliAuditLogRecord struct {

	// Arguments of a GPFS command
	Arguments string `json:"arguments,omitempty"`

	// Name of a GPFS command
	Command string `json:"command,omitempty"`

	// Entry Time of a GPFS command
	EntryTime string `json:"entryTime,omitempty"`

	// Exit Time of a GPFS command
	ExitTime string `json:"exitTime,omitempty"`

	// Name of the node where a GPFS command was running
	Node string `json:"node,omitempty"`

	// Originator of a GPFS command
	Originator string `json:"originator,omitempty"`

	// PID of a GPFS command
	Pid int32 `json:"pid,omitempty"`

	// Return Code of a GPFS command
	ReturnCode int32 `json:"returnCode,omitempty"`

	// User triggered a GPFS command
	User string `json:"user,omitempty"`
}

// Validate validates this cli audit log record
func (m *CliAuditLogRecord) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *CliAuditLogRecord) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CliAuditLogRecord) UnmarshalBinary(b []byte) error {
	var res CliAuditLogRecord
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
