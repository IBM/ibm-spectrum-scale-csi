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

	"example.com/m/v2/models"
)

// NewAcPutEntryv2Params creates a new AcPutEntryv2Params object
// with the default values initialized.
func NewAcPutEntryv2Params() *AcPutEntryv2Params {
	var ()
	return &AcPutEntryv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewAcPutEntryv2ParamsWithTimeout creates a new AcPutEntryv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewAcPutEntryv2ParamsWithTimeout(timeout time.Duration) *AcPutEntryv2Params {
	var ()
	return &AcPutEntryv2Params{

		timeout: timeout,
	}
}

// NewAcPutEntryv2ParamsWithContext creates a new AcPutEntryv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewAcPutEntryv2ParamsWithContext(ctx context.Context) *AcPutEntryv2Params {
	var ()
	return &AcPutEntryv2Params{

		Context: ctx,
	}
}

// NewAcPutEntryv2ParamsWithHTTPClient creates a new AcPutEntryv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewAcPutEntryv2ParamsWithHTTPClient(client *http.Client) *AcPutEntryv2Params {
	var ()
	return &AcPutEntryv2Params{
		HTTPClient: client,
	}
}

/*AcPutEntryv2Params contains all the parameters to send to the API endpoint
for the ac put entryv2 operation typically these are written to a http.Request
*/
type AcPutEntryv2Params struct {

	/*Body*/
	Body *models.SmbExportACLEntry
	/*Name
	  The name of the user,group or system

	*/
	Name string
	/*ShareName
	  The name of the smb share

	*/
	ShareName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the ac put entryv2 params
func (o *AcPutEntryv2Params) WithTimeout(timeout time.Duration) *AcPutEntryv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the ac put entryv2 params
func (o *AcPutEntryv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the ac put entryv2 params
func (o *AcPutEntryv2Params) WithContext(ctx context.Context) *AcPutEntryv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the ac put entryv2 params
func (o *AcPutEntryv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the ac put entryv2 params
func (o *AcPutEntryv2Params) WithHTTPClient(client *http.Client) *AcPutEntryv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the ac put entryv2 params
func (o *AcPutEntryv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the ac put entryv2 params
func (o *AcPutEntryv2Params) WithBody(body *models.SmbExportACLEntry) *AcPutEntryv2Params {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the ac put entryv2 params
func (o *AcPutEntryv2Params) SetBody(body *models.SmbExportACLEntry) {
	o.Body = body
}

// WithName adds the name to the ac put entryv2 params
func (o *AcPutEntryv2Params) WithName(name string) *AcPutEntryv2Params {
	o.SetName(name)
	return o
}

// SetName adds the name to the ac put entryv2 params
func (o *AcPutEntryv2Params) SetName(name string) {
	o.Name = name
}

// WithShareName adds the shareName to the ac put entryv2 params
func (o *AcPutEntryv2Params) WithShareName(shareName string) *AcPutEntryv2Params {
	o.SetShareName(shareName)
	return o
}

// SetShareName adds the shareName to the ac put entryv2 params
func (o *AcPutEntryv2Params) SetShareName(shareName string) {
	o.ShareName = shareName
}

// WriteToRequest writes these params to a swagger request
func (o *AcPutEntryv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	// path param shareName
	if err := r.SetPathParam("shareName", o.ShareName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
