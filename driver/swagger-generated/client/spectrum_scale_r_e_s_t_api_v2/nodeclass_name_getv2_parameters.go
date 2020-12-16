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

// NewNodeclassNameGetv2Params creates a new NodeclassNameGetv2Params object
// with the default values initialized.
func NewNodeclassNameGetv2Params() *NodeclassNameGetv2Params {
	var ()
	return &NodeclassNameGetv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewNodeclassNameGetv2ParamsWithTimeout creates a new NodeclassNameGetv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewNodeclassNameGetv2ParamsWithTimeout(timeout time.Duration) *NodeclassNameGetv2Params {
	var ()
	return &NodeclassNameGetv2Params{

		timeout: timeout,
	}
}

// NewNodeclassNameGetv2ParamsWithContext creates a new NodeclassNameGetv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewNodeclassNameGetv2ParamsWithContext(ctx context.Context) *NodeclassNameGetv2Params {
	var ()
	return &NodeclassNameGetv2Params{

		Context: ctx,
	}
}

// NewNodeclassNameGetv2ParamsWithHTTPClient creates a new NodeclassNameGetv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewNodeclassNameGetv2ParamsWithHTTPClient(client *http.Client) *NodeclassNameGetv2Params {
	var ()
	return &NodeclassNameGetv2Params{
		HTTPClient: client,
	}
}

/*NodeclassNameGetv2Params contains all the parameters to send to the API endpoint
for the nodeclass name getv2 operation typically these are written to a http.Request
*/
type NodeclassNameGetv2Params struct {

	/*Fields
	  Comma separated list of fields to be included in response. ':all:' selects all available fields.

	*/
	Fields *string
	/*Filter
	  Filter objects by expression, e.g. 'type=SYSTEM'

	*/
	Filter *string
	/*NodeclassName
	  nodeclass name

	*/
	NodeclassName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) WithTimeout(timeout time.Duration) *NodeclassNameGetv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) WithContext(ctx context.Context) *NodeclassNameGetv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) WithHTTPClient(client *http.Client) *NodeclassNameGetv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFields adds the fields to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) WithFields(fields *string) *NodeclassNameGetv2Params {
	o.SetFields(fields)
	return o
}

// SetFields adds the fields to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) SetFields(fields *string) {
	o.Fields = fields
}

// WithFilter adds the filter to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) WithFilter(filter *string) *NodeclassNameGetv2Params {
	o.SetFilter(filter)
	return o
}

// SetFilter adds the filter to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) SetFilter(filter *string) {
	o.Filter = filter
}

// WithNodeclassName adds the nodeclassName to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) WithNodeclassName(nodeclassName string) *NodeclassNameGetv2Params {
	o.SetNodeclassName(nodeclassName)
	return o
}

// SetNodeclassName adds the nodeclassName to the nodeclass name getv2 params
func (o *NodeclassNameGetv2Params) SetNodeclassName(nodeclassName string) {
	o.NodeclassName = nodeclassName
}

// WriteToRequest writes these params to a swagger request
func (o *NodeclassNameGetv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param nodeclassName
	if err := r.SetPathParam("nodeclassName", o.NodeclassName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
