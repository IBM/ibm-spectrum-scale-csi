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

// NewRemoteAccessPostParams creates a new RemoteAccessPostParams object
// with the default values initialized.
func NewRemoteAccessPostParams() *RemoteAccessPostParams {
	var ()
	return &RemoteAccessPostParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewRemoteAccessPostParamsWithTimeout creates a new RemoteAccessPostParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewRemoteAccessPostParamsWithTimeout(timeout time.Duration) *RemoteAccessPostParams {
	var ()
	return &RemoteAccessPostParams{

		timeout: timeout,
	}
}

// NewRemoteAccessPostParamsWithContext creates a new RemoteAccessPostParams object
// with the default values initialized, and the ability to set a context for a request
func NewRemoteAccessPostParamsWithContext(ctx context.Context) *RemoteAccessPostParams {
	var ()
	return &RemoteAccessPostParams{

		Context: ctx,
	}
}

// NewRemoteAccessPostParamsWithHTTPClient creates a new RemoteAccessPostParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewRemoteAccessPostParamsWithHTTPClient(client *http.Client) *RemoteAccessPostParams {
	var ()
	return &RemoteAccessPostParams{
		HTTPClient: client,
	}
}

/*RemoteAccessPostParams contains all the parameters to send to the API endpoint
for the remote access post operation typically these are written to a http.Request
*/
type RemoteAccessPostParams struct {

	/*Body
	  The data of the access request.

	*/
	Body *models.AccessRequestData

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the remote access post params
func (o *RemoteAccessPostParams) WithTimeout(timeout time.Duration) *RemoteAccessPostParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the remote access post params
func (o *RemoteAccessPostParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the remote access post params
func (o *RemoteAccessPostParams) WithContext(ctx context.Context) *RemoteAccessPostParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the remote access post params
func (o *RemoteAccessPostParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the remote access post params
func (o *RemoteAccessPostParams) WithHTTPClient(client *http.Client) *RemoteAccessPostParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the remote access post params
func (o *RemoteAccessPostParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the remote access post params
func (o *RemoteAccessPostParams) WithBody(body *models.AccessRequestData) *RemoteAccessPostParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the remote access post params
func (o *RemoteAccessPostParams) SetBody(body *models.AccessRequestData) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *RemoteAccessPostParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
