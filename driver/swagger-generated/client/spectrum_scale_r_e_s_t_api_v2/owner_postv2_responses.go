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

// OwnerPostv2Reader is a Reader for the OwnerPostv2 structure.
type OwnerPostv2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *OwnerPostv2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewOwnerPostv2Accepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewOwnerPostv2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewOwnerPostv2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewOwnerPostv2Accepted creates a OwnerPostv2Accepted with default headers values
func NewOwnerPostv2Accepted() *OwnerPostv2Accepted {
	return &OwnerPostv2Accepted{}
}

/*OwnerPostv2Accepted handles this case with default header values.

successful operation
*/
type OwnerPostv2Accepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *OwnerPostv2Accepted) Error() string {
	return fmt.Sprintf("[PUT /scalemgmt/v2/filesystems/{filesystemName}/owner/{path}][%d] ownerPostv2Accepted  %+v", 202, o.Payload)
}

func (o *OwnerPostv2Accepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *OwnerPostv2Accepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewOwnerPostv2BadRequest creates a OwnerPostv2BadRequest with default headers values
func NewOwnerPostv2BadRequest() *OwnerPostv2BadRequest {
	return &OwnerPostv2BadRequest{}
}

/*OwnerPostv2BadRequest handles this case with default header values.

Invalid fs, path or owner
*/
type OwnerPostv2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *OwnerPostv2BadRequest) Error() string {
	return fmt.Sprintf("[PUT /scalemgmt/v2/filesystems/{filesystemName}/owner/{path}][%d] ownerPostv2BadRequest  %+v", 400, o.Payload)
}

func (o *OwnerPostv2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *OwnerPostv2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewOwnerPostv2InternalServerError creates a OwnerPostv2InternalServerError with default headers values
func NewOwnerPostv2InternalServerError() *OwnerPostv2InternalServerError {
	return &OwnerPostv2InternalServerError{}
}

/*OwnerPostv2InternalServerError handles this case with default header values.

An unexpected error occurred.
*/
type OwnerPostv2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *OwnerPostv2InternalServerError) Error() string {
	return fmt.Sprintf("[PUT /scalemgmt/v2/filesystems/{filesystemName}/owner/{path}][%d] ownerPostv2InternalServerError  %+v", 500, o.Payload)
}

func (o *OwnerPostv2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *OwnerPostv2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
