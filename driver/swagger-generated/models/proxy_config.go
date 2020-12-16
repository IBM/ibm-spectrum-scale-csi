// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProxyConfig proxy config
//
// swagger:model ProxyConfig
type ProxyConfig struct {

	// Specifies if the proxy is enabled
	Enabled bool `json:"enabled,omitempty"`

	// The hostname of the proxy
	Location string `json:"location,omitempty"`

	// Only used for configuration. Will never be returned.
	Password string `json:"password,omitempty"`

	// The port of the proxy
	Port int32 `json:"port,omitempty"`

	// The username that is used to connect to the proxy
	Username string `json:"username,omitempty"`
}

// Validate validates this proxy config
func (m *ProxyConfig) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProxyConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProxyConfig) UnmarshalBinary(b []byte) error {
	var res ProxyConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
