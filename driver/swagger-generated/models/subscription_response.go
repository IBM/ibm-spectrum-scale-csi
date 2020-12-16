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

// SubscriptionResponse subscription response
//
// swagger:model SubscriptionResponse
type SubscriptionResponse struct {

	// Type of the subscription feed
	// Required: true
	FeedType *string `json:"feedType"`

	// URL of the subscription feed
	// Required: true
	FeedURL *string `json:"feedURL"`
}

// Validate validates this subscription response
func (m *SubscriptionResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFeedType(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateFeedURL(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SubscriptionResponse) validateFeedType(formats strfmt.Registry) error {

	if err := validate.Required("feedType", "body", m.FeedType); err != nil {
		return err
	}

	return nil
}

func (m *SubscriptionResponse) validateFeedURL(formats strfmt.Registry) error {

	if err := validate.Required("feedURL", "body", m.FeedURL); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *SubscriptionResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SubscriptionResponse) UnmarshalBinary(b []byte) error {
	var res SubscriptionResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
