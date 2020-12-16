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

// ElementGetReader is a Reader for the ElementGet structure.
type ElementGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ElementGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewElementGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewElementGetBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewElementGetInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewElementGetOK creates a ElementGetOK with default headers values
func NewElementGetOK() *ElementGetOK {
	return &ElementGetOK{}
}

/*ElementGetOK handles this case with default header values.

successful operation
*/
type ElementGetOK struct {
	Payload *models.ComponentElementInlineResponse200
}

func (o *ElementGetOK) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/components/elements/{componentId}][%d] elementGetOK  %+v", 200, o.Payload)
}

func (o *ElementGetOK) GetPayload() *models.ComponentElementInlineResponse200 {
	return o.Payload
}

func (o *ElementGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ComponentElementInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewElementGetBadRequest creates a ElementGetBadRequest with default headers values
func NewElementGetBadRequest() *ElementGetBadRequest {
	return &ElementGetBadRequest{}
}

/*ElementGetBadRequest handles this case with default header values.

Invalid request
*/
type ElementGetBadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *ElementGetBadRequest) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/components/elements/{componentId}][%d] elementGetBadRequest  %+v", 400, o.Payload)
}

func (o *ElementGetBadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *ElementGetBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewElementGetInternalServerError creates a ElementGetInternalServerError with default headers values
func NewElementGetInternalServerError() *ElementGetInternalServerError {
	return &ElementGetInternalServerError{}
}

/*ElementGetInternalServerError handles this case with default header values.

Internal Server Error
*/
type ElementGetInternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *ElementGetInternalServerError) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/components/elements/{componentId}][%d] elementGetInternalServerError  %+v", 500, o.Payload)
}

func (o *ElementGetInternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *ElementGetInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
