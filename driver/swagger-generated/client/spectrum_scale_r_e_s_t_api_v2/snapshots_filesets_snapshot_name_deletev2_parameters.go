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

// NewSnapshotsFilesetsSnapshotNameDeletev2Params creates a new SnapshotsFilesetsSnapshotNameDeletev2Params object
// with the default values initialized.
func NewSnapshotsFilesetsSnapshotNameDeletev2Params() *SnapshotsFilesetsSnapshotNameDeletev2Params {
	var ()
	return &SnapshotsFilesetsSnapshotNameDeletev2Params{

		timeout: cr.DefaultTimeout,
	}
}

// NewSnapshotsFilesetsSnapshotNameDeletev2ParamsWithTimeout creates a new SnapshotsFilesetsSnapshotNameDeletev2Params object
// with the default values initialized, and the ability to set a timeout on a request
func NewSnapshotsFilesetsSnapshotNameDeletev2ParamsWithTimeout(timeout time.Duration) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	var ()
	return &SnapshotsFilesetsSnapshotNameDeletev2Params{

		timeout: timeout,
	}
}

// NewSnapshotsFilesetsSnapshotNameDeletev2ParamsWithContext creates a new SnapshotsFilesetsSnapshotNameDeletev2Params object
// with the default values initialized, and the ability to set a context for a request
func NewSnapshotsFilesetsSnapshotNameDeletev2ParamsWithContext(ctx context.Context) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	var ()
	return &SnapshotsFilesetsSnapshotNameDeletev2Params{

		Context: ctx,
	}
}

// NewSnapshotsFilesetsSnapshotNameDeletev2ParamsWithHTTPClient creates a new SnapshotsFilesetsSnapshotNameDeletev2Params object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewSnapshotsFilesetsSnapshotNameDeletev2ParamsWithHTTPClient(client *http.Client) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	var ()
	return &SnapshotsFilesetsSnapshotNameDeletev2Params{
		HTTPClient: client,
	}
}

/*SnapshotsFilesetsSnapshotNameDeletev2Params contains all the parameters to send to the API endpoint
for the snapshots filesets snapshot name deletev2 operation typically these are written to a http.Request
*/
type SnapshotsFilesetsSnapshotNameDeletev2Params struct {

	/*FilesetName
	  The fileset name

	*/
	FilesetName string
	/*FilesystemName
	  The filesystem name, :all:, :all\_local: or :all\_remote:

	*/
	FilesystemName string
	/*SnapshotName
	  The snapshot name

	*/
	SnapshotName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WithTimeout(timeout time.Duration) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WithContext(ctx context.Context) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WithHTTPClient(client *http.Client) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFilesetName adds the filesetName to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WithFilesetName(filesetName string) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	o.SetFilesetName(filesetName)
	return o
}

// SetFilesetName adds the filesetName to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) SetFilesetName(filesetName string) {
	o.FilesetName = filesetName
}

// WithFilesystemName adds the filesystemName to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WithFilesystemName(filesystemName string) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	o.SetFilesystemName(filesystemName)
	return o
}

// SetFilesystemName adds the filesystemName to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) SetFilesystemName(filesystemName string) {
	o.FilesystemName = filesystemName
}

// WithSnapshotName adds the snapshotName to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WithSnapshotName(snapshotName string) *SnapshotsFilesetsSnapshotNameDeletev2Params {
	o.SetSnapshotName(snapshotName)
	return o
}

// SetSnapshotName adds the snapshotName to the snapshots filesets snapshot name deletev2 params
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) SetSnapshotName(snapshotName string) {
	o.SnapshotName = snapshotName
}

// WriteToRequest writes these params to a swagger request
func (o *SnapshotsFilesetsSnapshotNameDeletev2Params) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param filesetName
	if err := r.SetPathParam("filesetName", o.FilesetName); err != nil {
		return err
	}

	// path param filesystemName
	if err := r.SetPathParam("filesystemName", o.FilesystemName); err != nil {
		return err
	}

	// path param snapshotName
	if err := r.SetPathParam("snapshotName", o.SnapshotName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
