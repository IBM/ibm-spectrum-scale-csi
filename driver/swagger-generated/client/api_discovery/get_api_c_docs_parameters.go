// Code generated by go-swagger; DO NOT EDIT.

package api_discovery

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
	"github.com/go-openapi/swag"
)

// NewGetAPICDocsParams creates a new GetAPICDocsParams object
// with the default values initialized.
func NewGetAPICDocsParams() *GetAPICDocsParams {
	var ()
	return &GetAPICDocsParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPICDocsParamsWithTimeout creates a new GetAPICDocsParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetAPICDocsParamsWithTimeout(timeout time.Duration) *GetAPICDocsParams {
	var ()
	return &GetAPICDocsParams{

		timeout: timeout,
	}
}

// NewGetAPICDocsParamsWithContext creates a new GetAPICDocsParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetAPICDocsParamsWithContext(ctx context.Context) *GetAPICDocsParams {
	var ()
	return &GetAPICDocsParams{

		Context: ctx,
	}
}

// NewGetAPICDocsParamsWithHTTPClient creates a new GetAPICDocsParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetAPICDocsParamsWithHTTPClient(client *http.Client) *GetAPICDocsParams {
	var ()
	return &GetAPICDocsParams{
		HTTPClient: client,
	}
}

/*GetAPICDocsParams contains all the parameters to send to the API endpoint
for the get API c docs operation typically these are written to a http.Request
*/
type GetAPICDocsParams struct {

	/*Accept
	  Format of the returned document

	*/
	Accept *string
	/*Root
	  Filter the found context roots

	*/
	Root []string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get API c docs params
func (o *GetAPICDocsParams) WithTimeout(timeout time.Duration) *GetAPICDocsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API c docs params
func (o *GetAPICDocsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API c docs params
func (o *GetAPICDocsParams) WithContext(ctx context.Context) *GetAPICDocsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API c docs params
func (o *GetAPICDocsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API c docs params
func (o *GetAPICDocsParams) WithHTTPClient(client *http.Client) *GetAPICDocsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API c docs params
func (o *GetAPICDocsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAccept adds the accept to the get API c docs params
func (o *GetAPICDocsParams) WithAccept(accept *string) *GetAPICDocsParams {
	o.SetAccept(accept)
	return o
}

// SetAccept adds the accept to the get API c docs params
func (o *GetAPICDocsParams) SetAccept(accept *string) {
	o.Accept = accept
}

// WithRoot adds the root to the get API c docs params
func (o *GetAPICDocsParams) WithRoot(root []string) *GetAPICDocsParams {
	o.SetRoot(root)
	return o
}

// SetRoot adds the root to the get API c docs params
func (o *GetAPICDocsParams) SetRoot(root []string) {
	o.Root = root
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPICDocsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Accept != nil {

		// header param accept
		if err := r.SetHeaderParam("accept", *o.Accept); err != nil {
			return err
		}

	}

	valuesRoot := o.Root

	joinedRoot := swag.JoinByFormat(valuesRoot, "multi")
	// query array param root
	if err := r.SetQueryParam("root", joinedRoot...); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
