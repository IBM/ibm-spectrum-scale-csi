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

// NewThresholdDeletev2Params creates a new ThresholdDeletev2Params object
// with the default values initialized.
func NewThresholdDeletev2Params() *ThresholdDeletev2Params {
	var ()
	return &ThresholdDeletev2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewThresholdDeletev2ParamsWithTimeout creates a new ThresholdDeletev2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewThresholdDeletev2ParamsWithTimeout(timeout time.Duration) *ThresholdDeletev2Params {
	var ()
	return &ThresholdDeletev2Params{

		timeout: timeout,
	}
}

// NewThresholdDeletev2ParamsWithContext creates a new ThresholdDeletev2Params object
// with the default values initialized, and the ability to set a context for a request
func NewThresholdDeletev2ParamsWithContext(ctx context.Context) *ThresholdDeletev2Params {
	var ()
	return &ThresholdDeletev2Params{

		Context: ctx,
	}
}

// NewThresholdDeletev2ParamsWithHTTPClient creates a new ThresholdDeletev2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewThresholdDeletev2ParamsWithHTTPClient(client *http.Client) *ThresholdDeletev2Params {
	var ()
	return &ThresholdDeletev2Params{
		HTTPClient: client,
	}
}

/*ThresholdDeletev2Params contains all the parameters to send to the API endpoint
for the threshold deletev2 operation typically these are written to a http.Request
*/
type ThresholdDeletev2Params struct {

	/*Name
	  threshold rule name

	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the threshold deletev2 params
func (o *ThresholdDeletev2Params) WithTimeout(timeout time.Duration) *ThresholdDeletev2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the threshold deletev2 params
func (o *ThresholdDeletev2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the threshold deletev2 params
func (o *ThresholdDeletev2Params) WithContext(ctx context.Context) *ThresholdDeletev2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the threshold deletev2 params
func (o *ThresholdDeletev2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the threshold deletev2 params
func (o *ThresholdDeletev2Params) WithHTTPClient(client *http.Client) *ThresholdDeletev2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the threshold deletev2 params
func (o *ThresholdDeletev2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the threshold deletev2 params
func (o *ThresholdDeletev2Params) WithName(name string) *ThresholdDeletev2Params {
	o.SetName(name)
	return o
}

// SetName adds the name to the threshold deletev2 params
func (o *ThresholdDeletev2Params) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *ThresholdDeletev2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
