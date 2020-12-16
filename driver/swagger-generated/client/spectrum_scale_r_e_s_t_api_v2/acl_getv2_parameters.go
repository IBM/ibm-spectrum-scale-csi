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

// NewACLGetv2Params creates a new ACLGetv2Params object
// with the default values initialized.
func NewACLGetv2Params() *ACLGetv2Params {
	var ()
	return &ACLGetv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewACLGetv2ParamsWithTimeout creates a new ACLGetv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewACLGetv2ParamsWithTimeout(timeout time.Duration) *ACLGetv2Params {
	var ()
	return &ACLGetv2Params{

		timeout: timeout,
	}
}

// NewACLGetv2ParamsWithContext creates a new ACLGetv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewACLGetv2ParamsWithContext(ctx context.Context) *ACLGetv2Params {
	var ()
	return &ACLGetv2Params{

		Context: ctx,
	}
}

// NewACLGetv2ParamsWithHTTPClient creates a new ACLGetv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewACLGetv2ParamsWithHTTPClient(client *http.Client) *ACLGetv2Params {
	var ()
	return &ACLGetv2Params{
		HTTPClient: client,
	}
}

/*ACLGetv2Params contains all the parameters to send to the API endpoint
for the acl getv2 operation typically these are written to a http.Request
*/
type ACLGetv2Params struct {

	/*Fields
	  Comma separated list of fields to be included in response. ':all:' selects all available fields.

	*/
	Fields *string
	/*FilesystemName
	  name of the filesystem

	*/
	FilesystemName string
	/*Path
	  The file path relative to the filesystem's mount point

	*/
	Path string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the acl getv2 params
func (o *ACLGetv2Params) WithTimeout(timeout time.Duration) *ACLGetv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the acl getv2 params
func (o *ACLGetv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the acl getv2 params
func (o *ACLGetv2Params) WithContext(ctx context.Context) *ACLGetv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the acl getv2 params
func (o *ACLGetv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the acl getv2 params
func (o *ACLGetv2Params) WithHTTPClient(client *http.Client) *ACLGetv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the acl getv2 params
func (o *ACLGetv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFields adds the fields to the acl getv2 params
func (o *ACLGetv2Params) WithFields(fields *string) *ACLGetv2Params {
	o.SetFields(fields)
	return o
}

// SetFields adds the fields to the acl getv2 params
func (o *ACLGetv2Params) SetFields(fields *string) {
	o.Fields = fields
}

// WithFilesystemName adds the filesystemName to the acl getv2 params
func (o *ACLGetv2Params) WithFilesystemName(filesystemName string) *ACLGetv2Params {
	o.SetFilesystemName(filesystemName)
	return o
}

// SetFilesystemName adds the filesystemName to the acl getv2 params
func (o *ACLGetv2Params) SetFilesystemName(filesystemName string) {
	o.FilesystemName = filesystemName
}

// WithPath adds the path to the acl getv2 params
func (o *ACLGetv2Params) WithPath(path string) *ACLGetv2Params {
	o.SetPath(path)
	return o
}

// SetPath adds the path to the acl getv2 params
func (o *ACLGetv2Params) SetPath(path string) {
	o.Path = path
}

// WriteToRequest writes these params to a swagger request
func (o *ACLGetv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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

	// path param path
	if err := r.SetPathParam("path", o.Path); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
