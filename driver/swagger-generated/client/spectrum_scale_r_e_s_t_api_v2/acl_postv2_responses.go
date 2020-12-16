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

// ACLPostv2Reader is a Reader for the ACLPostv2 structure.
type ACLPostv2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ACLPostv2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewACLPostv2Accepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewACLPostv2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewACLPostv2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewACLPostv2Accepted creates a ACLPostv2Accepted with default headers values
func NewACLPostv2Accepted() *ACLPostv2Accepted {
	return &ACLPostv2Accepted{}
}

/*ACLPostv2Accepted handles this case with default header values.

successful operation
*/
type ACLPostv2Accepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *ACLPostv2Accepted) Error() string {
	return fmt.Sprintf("[PUT /scalemgmt/v2/filesystems/{filesystemName}/acl/{path}][%d] aclPostv2Accepted  %+v", 202, o.Payload)
}

func (o *ACLPostv2Accepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *ACLPostv2Accepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewACLPostv2BadRequest creates a ACLPostv2BadRequest with default headers values
func NewACLPostv2BadRequest() *ACLPostv2BadRequest {
	return &ACLPostv2BadRequest{}
}

/*ACLPostv2BadRequest handles this case with default header values.

Invalid fs, path or acl
*/
type ACLPostv2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *ACLPostv2BadRequest) Error() string {
	return fmt.Sprintf("[PUT /scalemgmt/v2/filesystems/{filesystemName}/acl/{path}][%d] aclPostv2BadRequest  %+v", 400, o.Payload)
}

func (o *ACLPostv2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *ACLPostv2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewACLPostv2InternalServerError creates a ACLPostv2InternalServerError with default headers values
func NewACLPostv2InternalServerError() *ACLPostv2InternalServerError {
	return &ACLPostv2InternalServerError{}
}

/*ACLPostv2InternalServerError handles this case with default header values.

An unexpected error occurred.
*/
type ACLPostv2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *ACLPostv2InternalServerError) Error() string {
	return fmt.Sprintf("[PUT /scalemgmt/v2/filesystems/{filesystemName}/acl/{path}][%d] aclPostv2InternalServerError  %+v", 500, o.Payload)
}

func (o *ACLPostv2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *ACLPostv2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
