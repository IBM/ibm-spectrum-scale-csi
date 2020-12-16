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

// NewNodesListMappingv2Params creates a new NodesListMappingv2Params object
// with the default values initialized.
func NewNodesListMappingv2Params() *NodesListMappingv2Params {
	var ()
	return &NodesListMappingv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewNodesListMappingv2ParamsWithTimeout creates a new NodesListMappingv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewNodesListMappingv2ParamsWithTimeout(timeout time.Duration) *NodesListMappingv2Params {
	var ()
	return &NodesListMappingv2Params{

		timeout: timeout,
	}
}

// NewNodesListMappingv2ParamsWithContext creates a new NodesListMappingv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewNodesListMappingv2ParamsWithContext(ctx context.Context) *NodesListMappingv2Params {
	var ()
	return &NodesListMappingv2Params{

		Context: ctx,
	}
}

// NewNodesListMappingv2ParamsWithHTTPClient creates a new NodesListMappingv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewNodesListMappingv2ParamsWithHTTPClient(client *http.Client) *NodesListMappingv2Params {
	var ()
	return &NodesListMappingv2Params{
		HTTPClient: client,
	}
}

/*NodesListMappingv2Params contains all the parameters to send to the API endpoint
for the nodes list mappingv2 operation typically these are written to a http.Request
*/
type NodesListMappingv2Params struct {

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

// WithTimeout adds the timeout to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) WithTimeout(timeout time.Duration) *NodesListMappingv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) WithContext(ctx context.Context) *NodesListMappingv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) WithHTTPClient(client *http.Client) *NodesListMappingv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFields adds the fields to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) WithFields(fields *string) *NodesListMappingv2Params {
	o.SetFields(fields)
	return o
}

// SetFields adds the fields to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) SetFields(fields *string) {
	o.Fields = fields
}

// WithFilter adds the filter to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) WithFilter(filter *string) *NodesListMappingv2Params {
	o.SetFilter(filter)
	return o
}

// SetFilter adds the filter to the nodes list mappingv2 params
func (o *NodesListMappingv2Params) SetFilter(filter *string) {
	o.Filter = filter
}

// WriteToRequest writes these params to a swagger request
func (o *NodesListMappingv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
