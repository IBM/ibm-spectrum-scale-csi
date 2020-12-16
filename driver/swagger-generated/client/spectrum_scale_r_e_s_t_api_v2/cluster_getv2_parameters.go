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

// NewClusterGetv2Params creates a new ClusterGetv2Params object
// with the default values initialized.
func NewClusterGetv2Params() *ClusterGetv2Params {

	return &ClusterGetv2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewClusterGetv2ParamsWithTimeout creates a new ClusterGetv2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewClusterGetv2ParamsWithTimeout(timeout time.Duration) *ClusterGetv2Params {

	return &ClusterGetv2Params{

		timeout: timeout,
	}
}

// NewClusterGetv2ParamsWithContext creates a new ClusterGetv2Params object
// with the default values initialized, and the ability to set a context for a request
func NewClusterGetv2ParamsWithContext(ctx context.Context) *ClusterGetv2Params {

	return &ClusterGetv2Params{

		Context: ctx,
	}
}

// NewClusterGetv2ParamsWithHTTPClient creates a new ClusterGetv2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewClusterGetv2ParamsWithHTTPClient(client *http.Client) *ClusterGetv2Params {

	return &ClusterGetv2Params{
		HTTPClient: client,
	}
}

/*ClusterGetv2Params contains all the parameters to send to the API endpoint
for the cluster getv2 operation typically these are written to a http.Request
*/
type ClusterGetv2Params struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the cluster getv2 params
func (o *ClusterGetv2Params) WithTimeout(timeout time.Duration) *ClusterGetv2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the cluster getv2 params
func (o *ClusterGetv2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the cluster getv2 params
func (o *ClusterGetv2Params) WithContext(ctx context.Context) *ClusterGetv2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the cluster getv2 params
func (o *ClusterGetv2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the cluster getv2 params
func (o *ClusterGetv2Params) WithHTTPClient(client *http.Client) *ClusterGetv2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the cluster getv2 params
func (o *ClusterGetv2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ClusterGetv2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
