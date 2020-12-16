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

// NsdsNameGetv2Reader is a Reader for the NsdsNameGetv2 structure.
type NsdsNameGetv2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *NsdsNameGetv2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewNsdsNameGetv2OK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewNsdsNameGetv2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewNsdsNameGetv2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewNsdsNameGetv2OK creates a NsdsNameGetv2OK with default headers values
func NewNsdsNameGetv2OK() *NsdsNameGetv2OK {
	return &NsdsNameGetv2OK{}
}

/*NsdsNameGetv2OK handles this case with default header values.

successful operation
*/
type NsdsNameGetv2OK struct {
	Payload *models.NsdInlineResponse200
}

func (o *NsdsNameGetv2OK) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/nsds/{nsdName}][%d] nsdsNameGetv2OK  %+v", 200, o.Payload)
}

func (o *NsdsNameGetv2OK) GetPayload() *models.NsdInlineResponse200 {
	return o.Payload
}

func (o *NsdsNameGetv2OK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.NsdInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewNsdsNameGetv2BadRequest creates a NsdsNameGetv2BadRequest with default headers values
func NewNsdsNameGetv2BadRequest() *NsdsNameGetv2BadRequest {
	return &NsdsNameGetv2BadRequest{}
}

/*NsdsNameGetv2BadRequest handles this case with default header values.

Invalid request
*/
type NsdsNameGetv2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *NsdsNameGetv2BadRequest) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/nsds/{nsdName}][%d] nsdsNameGetv2BadRequest  %+v", 400, o.Payload)
}

func (o *NsdsNameGetv2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *NsdsNameGetv2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewNsdsNameGetv2InternalServerError creates a NsdsNameGetv2InternalServerError with default headers values
func NewNsdsNameGetv2InternalServerError() *NsdsNameGetv2InternalServerError {
	return &NsdsNameGetv2InternalServerError{}
}

/*NsdsNameGetv2InternalServerError handles this case with default header values.

Internal Server Error
*/
type NsdsNameGetv2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *NsdsNameGetv2InternalServerError) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/nsds/{nsdName}][%d] nsdsNameGetv2InternalServerError  %+v", 500, o.Payload)
}

func (o *NsdsNameGetv2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *NsdsNameGetv2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
