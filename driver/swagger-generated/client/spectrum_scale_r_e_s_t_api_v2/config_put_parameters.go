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

// NewConfigPutParams creates a new ConfigPutParams object
// with the default values initialized.
func NewConfigPutParams() *ConfigPutParams {
	var ()
	return &ConfigPutParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewConfigPutParamsWithTimeout creates a new ConfigPutParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewConfigPutParamsWithTimeout(timeout time.Duration) *ConfigPutParams {
	var ()
	return &ConfigPutParams{

		timeout: timeout,
	}
}

// NewConfigPutParamsWithContext creates a new ConfigPutParams object
// with the default values initialized, and the ability to set a context for a request
func NewConfigPutParamsWithContext(ctx context.Context) *ConfigPutParams {
	var ()
	return &ConfigPutParams{

		Context: ctx,
	}
}

// NewConfigPutParamsWithHTTPClient creates a new ConfigPutParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewConfigPutParamsWithHTTPClient(client *http.Client) *ConfigPutParams {
	var ()
	return &ConfigPutParams{
		HTTPClient: client,
	}
}

/*ConfigPutParams contains all the parameters to send to the API endpoint
for the config put operation typically these are written to a http.Request
*/
type ConfigPutParams struct {

	/*Body*/
	Body *models.ConfigUpdate

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the config put params
func (o *ConfigPutParams) WithTimeout(timeout time.Duration) *ConfigPutParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the config put params
func (o *ConfigPutParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the config put params
func (o *ConfigPutParams) WithContext(ctx context.Context) *ConfigPutParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the config put params
func (o *ConfigPutParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the config put params
func (o *ConfigPutParams) WithHTTPClient(client *http.Client) *ConfigPutParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the config put params
func (o *ConfigPutParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the config put params
func (o *ConfigPutParams) WithBody(body *models.ConfigUpdate) *ConfigPutParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the config put params
func (o *ConfigPutParams) SetBody(body *models.ConfigUpdate) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *ConfigPutParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
