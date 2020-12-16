// Code generated by go-swagger; DO NOT EDIT.

package spectrum_scale_r_e_s_t_api_v2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewCesaddressesGetv2Params creates a new CesaddressesGetv2Params object
// with the default values initialized.
func NewCesaddressesGetv2Params() *CesaddressesGetv2Params {
	var ()
	return &CesaddressesGetv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewCesaddressesGetv2ParamsWithTimeout creates a new CesaddressesGetv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewCesaddressesGetv2ParamsWithTimeout(timeout time.Duration) *CesaddressesGetv2Params {
	var ()
	return &CesaddressesGetv2Params{

		timeout: timeout,
	}
}

// NewCesaddressesGetv2ParamsWithContext creates a new CesaddressesGetv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewCesaddressesGetv2ParamsWithContext(ctx context.Context) *CesaddressesGetv2Params {
	var ()
	return &CesaddressesGetv2Params{

		Context: ctx,
	}
}

// NewCesaddressesGetv2ParamsWithHTTPClient creates a new CesaddressesGetv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCesaddressesGetv2ParamsWithHTTPClient(client *http.Client) *CesaddressesGetv2Params {
	var ()
	return &CesaddressesGetv2Params{
		HTTPClient: client,
	}
}

/*CesaddressesGetv2Params contains all the parameters to send to the API endpoint
for the cesaddresses getv2 operation typically these are written to a http.Request
*/
type CesaddressesGetv2Params struct {

	/*Fields
	  Comma separated list of fields to be included in response. ':all:' selects all available fields.

	*/
	Fields *string
	/*Filter
	  Filter objects by expression, e.g. 'status=HEALTHY,entityType=FILESET'

	*/
	Filter *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) WithTimeout(timeout time.Duration) *CesaddressesGetv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) WithContext(ctx context.Context) *CesaddressesGetv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) WithHTTPClient(client *http.Client) *CesaddressesGetv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFields adds the fields to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) WithFields(fields *string) *CesaddressesGetv2Params {
	o.SetFields(fields)
	return o
}

// SetFields adds the fields to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) SetFields(fields *string) {
	o.Fields = fields
}

// WithFilter adds the filter to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) WithFilter(filter *string) *CesaddressesGetv2Params {
	o.SetFilter(filter)
	return o
}

// SetFilter adds the filter to the cesaddresses getv2 params
func (o *CesaddressesGetv2Params) SetFilter(filter *string) {
	o.Filter = filter
}

// WriteToRequest writes these params to a swagger request
func (o *CesaddressesGetv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Fields != nil {

		// query param fields
		var qrFields string
		if o.Fields != nil {
			qrFields = *o.Fields
		}
		qFields := qrFields
		if qFields != "" {
			if err := r.SetQueryParam("fields", qFields); err != nil {
				return err
			}
		}

	}

	if o.Filter != nil {

		// query param filter
		var qrFilter string
		if o.Filter != nil {
			qrFilter = *o.Filter
		}
		qFilter := qrFilter
		if qFilter != "" {
			if err := r.SetQueryParam("filter", qFilter); err != nil {
				return err
			}
		}

	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
