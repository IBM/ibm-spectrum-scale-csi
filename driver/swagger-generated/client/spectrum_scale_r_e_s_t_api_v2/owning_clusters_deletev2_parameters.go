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

// NewOwningClustersDeletev2Params creates a new OwningClustersDeletev2Params object
// with the default values initialized.
func NewOwningClustersDeletev2Params() *OwningClustersDeletev2Params {
	var ()
	return &OwningClustersDeletev2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewOwningClustersDeletev2ParamsWithTimeout creates a new OwningClustersDeletev2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewOwningClustersDeletev2ParamsWithTimeout(timeout time.Duration) *OwningClustersDeletev2Params {
	var ()
	return &OwningClustersDeletev2Params{

		timeout: timeout,
	}
}

// NewOwningClustersDeletev2ParamsWithContext creates a new OwningClustersDeletev2Params object
// with the default values initialized, and the ability to set a context for a request
func NewOwningClustersDeletev2ParamsWithContext(ctx context.Context) *OwningClustersDeletev2Params {
	var ()
	return &OwningClustersDeletev2Params{

		Context: ctx,
	}
}

// NewOwningClustersDeletev2ParamsWithHTTPClient creates a new OwningClustersDeletev2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewOwningClustersDeletev2ParamsWithHTTPClient(client *http.Client) *OwningClustersDeletev2Params {
	var ()
	return &OwningClustersDeletev2Params{
		HTTPClient: client,
	}
}

/*OwningClustersDeletev2Params contains all the parameters to send to the API endpoint
for the owning clusters deletev2 operation typically these are written to a http.Request
*/
type OwningClustersDeletev2Params struct {

	/*OwningCluster
	  owning cluster name

	*/
	OwningCluster string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) WithTimeout(timeout time.Duration) *OwningClustersDeletev2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) WithContext(ctx context.Context) *OwningClustersDeletev2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) WithHTTPClient(client *http.Client) *OwningClustersDeletev2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithOwningCluster adds the owningCluster to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) WithOwningCluster(owningCluster string) *OwningClustersDeletev2Params {
	o.SetOwningCluster(owningCluster)
	return o
}

// SetOwningCluster adds the owningCluster to the owning clusters deletev2 params
func (o *OwningClustersDeletev2Params) SetOwningCluster(owningCluster string) {
	o.OwningCluster = owningCluster
}

// WriteToRequest writes these params to a swagger request
func (o *OwningClustersDeletev2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param owningCluster
	if err := r.SetPathParam("owningCluster", o.OwningCluster); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
