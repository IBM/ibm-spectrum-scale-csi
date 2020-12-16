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

// NewDisksGetv2Params creates a new DisksGetv2Params object
// with the default values initialized.
func NewDisksGetv2Params() *DisksGetv2Params {
	var ()
	return &DisksGetv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewDisksGetv2ParamsWithTimeout creates a new DisksGetv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewDisksGetv2ParamsWithTimeout(timeout time.Duration) *DisksGetv2Params {
	var ()
	return &DisksGetv2Params{

		timeout: timeout,
	}
}

// NewDisksGetv2ParamsWithContext creates a new DisksGetv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewDisksGetv2ParamsWithContext(ctx context.Context) *DisksGetv2Params {
	var ()
	return &DisksGetv2Params{

		Context: ctx,
	}
}

// NewDisksGetv2ParamsWithHTTPClient creates a new DisksGetv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewDisksGetv2ParamsWithHTTPClient(client *http.Client) *DisksGetv2Params {
	var ()
	return &DisksGetv2Params{
		HTTPClient: client,
	}
}

/*DisksGetv2Params contains all the parameters to send to the API endpoint
for the disks getv2 operation typically these are written to a http.Request
*/
type DisksGetv2Params struct {

	/*Fields
	  Comma separated list of fields to be included in response. ':all:' selects all available fields.

	*/
	Fields *string
	/*FilesystemName
	  The filesystem name, :all:, :all\_local: or :all\_remote:

	*/
	FilesystemName string
	/*Filter
	  Filter objects by expression, e.g. 'status=HEALTHY,entityType=FILESET'

	*/
	Filter *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the disks getv2 params
func (o *DisksGetv2Params) WithTimeout(timeout time.Duration) *DisksGetv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the disks getv2 params
func (o *DisksGetv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the disks getv2 params
func (o *DisksGetv2Params) WithContext(ctx context.Context) *DisksGetv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the disks getv2 params
func (o *DisksGetv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the disks getv2 params
func (o *DisksGetv2Params) WithHTTPClient(client *http.Client) *DisksGetv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the disks getv2 params
func (o *DisksGetv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFields adds the fields to the disks getv2 params
func (o *DisksGetv2Params) WithFields(fields *string) *DisksGetv2Params {
	o.SetFields(fields)
	return o
}

// SetFields adds the fields to the disks getv2 params
func (o *DisksGetv2Params) SetFields(fields *string) {
	o.Fields = fields
}

// WithFilesystemName adds the filesystemName to the disks getv2 params
func (o *DisksGetv2Params) WithFilesystemName(filesystemName string) *DisksGetv2Params {
	o.SetFilesystemName(filesystemName)
	return o
}

// SetFilesystemName adds the filesystemName to the disks getv2 params
func (o *DisksGetv2Params) SetFilesystemName(filesystemName string) {
	o.FilesystemName = filesystemName
}

// WithFilter adds the filter to the disks getv2 params
func (o *DisksGetv2Params) WithFilter(filter *string) *DisksGetv2Params {
	o.SetFilter(filter)
	return o
}

// SetFilter adds the filter to the disks getv2 params
func (o *DisksGetv2Params) SetFilter(filter *string) {
	o.Filter = filter
}

// WriteToRequest writes these params to a swagger request
func (o *DisksGetv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param filesystemName
	if err := r.SetPathParam("filesystemName", o.FilesystemName); err != nil {
		return err
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
