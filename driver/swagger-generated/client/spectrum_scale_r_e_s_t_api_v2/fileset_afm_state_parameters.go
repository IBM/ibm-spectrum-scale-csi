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

// NewFilesetAfmStateParams creates a new FilesetAfmStateParams object
// with the default values initialized.
func NewFilesetAfmStateParams() *FilesetAfmStateParams {
	var ()
	return &FilesetAfmStateParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewFilesetAfmStateParamsWithTimeout creates a new FilesetAfmStateParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewFilesetAfmStateParamsWithTimeout(timeout time.Duration) *FilesetAfmStateParams {
	var ()
	return &FilesetAfmStateParams{

		timeout: timeout,
	}
}

// NewFilesetAfmStateParamsWithContext creates a new FilesetAfmStateParams object
// with the default values initialized, and the ability to set a context for a request
func NewFilesetAfmStateParamsWithContext(ctx context.Context) *FilesetAfmStateParams {
	var ()
	return &FilesetAfmStateParams{

		Context: ctx,
	}
}

// NewFilesetAfmStateParamsWithHTTPClient creates a new FilesetAfmStateParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewFilesetAfmStateParamsWithHTTPClient(client *http.Client) *FilesetAfmStateParams {
	var ()
	return &FilesetAfmStateParams{
		HTTPClient: client,
	}
}

/*FilesetAfmStateParams contains all the parameters to send to the API endpoint
for the fileset afm state operation typically these are written to a http.Request
*/
type FilesetAfmStateParams struct {

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

// WithTimeout adds the timeout to the fileset afm state params
func (o *FilesetAfmStateParams) WithTimeout(timeout time.Duration) *FilesetAfmStateParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the fileset afm state params
func (o *FilesetAfmStateParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the fileset afm state params
func (o *FilesetAfmStateParams) WithContext(ctx context.Context) *FilesetAfmStateParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the fileset afm state params
func (o *FilesetAfmStateParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the fileset afm state params
func (o *FilesetAfmStateParams) WithHTTPClient(client *http.Client) *FilesetAfmStateParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the fileset afm state params
func (o *FilesetAfmStateParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFields adds the fields to the fileset afm state params
func (o *FilesetAfmStateParams) WithFields(fields *string) *FilesetAfmStateParams {
	o.SetFields(fields)
	return o
}

// SetFields adds the fields to the fileset afm state params
func (o *FilesetAfmStateParams) SetFields(fields *string) {
	o.Fields = fields
}

// WithFilesystemName adds the filesystemName to the fileset afm state params
func (o *FilesetAfmStateParams) WithFilesystemName(filesystemName string) *FilesetAfmStateParams {
	o.SetFilesystemName(filesystemName)
	return o
}

// SetFilesystemName adds the filesystemName to the fileset afm state params
func (o *FilesetAfmStateParams) SetFilesystemName(filesystemName string) {
	o.FilesystemName = filesystemName
}

// WithFilter adds the filter to the fileset afm state params
func (o *FilesetAfmStateParams) WithFilter(filter *string) *FilesetAfmStateParams {
	o.SetFilter(filter)
	return o
}

// SetFilter adds the filter to the fileset afm state params
func (o *FilesetAfmStateParams) SetFilter(filter *string) {
	o.Filter = filter
}

// WriteToRequest writes these params to a swagger request
func (o *FilesetAfmStateParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
