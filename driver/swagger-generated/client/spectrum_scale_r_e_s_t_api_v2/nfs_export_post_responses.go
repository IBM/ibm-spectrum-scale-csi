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

// NfsExportPostReader is a Reader for the NfsExportPost structure.
type NfsExportPostReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *NfsExportPostReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewNfsExportPostAccepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewNfsExportPostInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewNfsExportPostAccepted creates a NfsExportPostAccepted with default headers values
func NewNfsExportPostAccepted() *NfsExportPostAccepted {
	return &NfsExportPostAccepted{}
}

/*NfsExportPostAccepted handles this case with default header values.

successful operation
*/
type NfsExportPostAccepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *NfsExportPostAccepted) Error() string {
	return fmt.Sprintf("[POST /scalemgmt/v2/nfs/exports][%d] nfsExportPostAccepted  %+v", 202, o.Payload)
}

func (o *NfsExportPostAccepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *NfsExportPostAccepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewNfsExportPostInternalServerError creates a NfsExportPostInternalServerError with default headers values
func NewNfsExportPostInternalServerError() *NfsExportPostInternalServerError {
	return &NfsExportPostInternalServerError{}
}

/*NfsExportPostInternalServerError handles this case with default header values.

Internal Server Error
*/
type NfsExportPostInternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *NfsExportPostInternalServerError) Error() string {
	return fmt.Sprintf("[POST /scalemgmt/v2/nfs/exports][%d] nfsExportPostInternalServerError  %+v", 500, o.Payload)
}

func (o *NfsExportPostInternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *NfsExportPostInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
