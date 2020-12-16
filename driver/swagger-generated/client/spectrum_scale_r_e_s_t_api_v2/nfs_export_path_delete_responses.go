// Code generated by go-swagger; DO NOT EDIT.

package spectrum_scale_r_e_s_t_api_v2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"example.com/m/v2/models"
)

// NfsExportPathDeleteReader is a Reader for the NfsExportPathDelete structure.
type NfsExportPathDeleteReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *NfsExportPathDeleteReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewNfsExportPathDeleteAccepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewNfsExportPathDeleteBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewNfsExportPathDeleteInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewNfsExportPathDeleteAccepted creates a NfsExportPathDeleteAccepted with default headers values
func NewNfsExportPathDeleteAccepted() *NfsExportPathDeleteAccepted {
	return &NfsExportPathDeleteAccepted{}
}

/*NfsExportPathDeleteAccepted handles this case with default header values.

successful operation
*/
type NfsExportPathDeleteAccepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *NfsExportPathDeleteAccepted) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/nfs/exports/{exportPath}][%d] nfsExportPathDeleteAccepted  %+v", 202, o.Payload)
}

func (o *NfsExportPathDeleteAccepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *NfsExportPathDeleteAccepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewNfsExportPathDeleteBadRequest creates a NfsExportPathDeleteBadRequest with default headers values
func NewNfsExportPathDeleteBadRequest() *NfsExportPathDeleteBadRequest {
	return &NfsExportPathDeleteBadRequest{}
}

/*NfsExportPathDeleteBadRequest handles this case with default header values.

NFS export not found
*/
type NfsExportPathDeleteBadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *NfsExportPathDeleteBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/nfs/exports/{exportPath}][%d] nfsExportPathDeleteBadRequest  %+v", 400, o.Payload)
}

func (o *NfsExportPathDeleteBadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *NfsExportPathDeleteBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewNfsExportPathDeleteInternalServerError creates a NfsExportPathDeleteInternalServerError with default headers values
func NewNfsExportPathDeleteInternalServerError() *NfsExportPathDeleteInternalServerError {
	return &NfsExportPathDeleteInternalServerError{}
}

/*NfsExportPathDeleteInternalServerError handles this case with default header values.

Internal Server Error
*/
type NfsExportPathDeleteInternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *NfsExportPathDeleteInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/nfs/exports/{exportPath}][%d] nfsExportPathDeleteInternalServerError  %+v", 500, o.Payload)
}

func (o *NfsExportPathDeleteInternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *NfsExportPathDeleteInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
