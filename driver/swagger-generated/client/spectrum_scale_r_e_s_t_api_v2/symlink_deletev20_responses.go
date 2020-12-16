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

// SymlinkDeletev20Reader is a Reader for the SymlinkDeletev20 structure.
type SymlinkDeletev20Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *SymlinkDeletev20Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewSymlinkDeletev20Accepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewSymlinkDeletev20BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewSymlinkDeletev20InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewSymlinkDeletev20Accepted creates a SymlinkDeletev20Accepted with default headers values
func NewSymlinkDeletev20Accepted() *SymlinkDeletev20Accepted {
	return &SymlinkDeletev20Accepted{}
}

/*SymlinkDeletev20Accepted handles this case with default header values.

successful operation
*/
type SymlinkDeletev20Accepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *SymlinkDeletev20Accepted) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/filesystems/{filesystemName}/filesets/{filesetName}/symlink/{path}][%d] symlinkDeletev20Accepted  %+v", 202, o.Payload)
}

func (o *SymlinkDeletev20Accepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *SymlinkDeletev20Accepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSymlinkDeletev20BadRequest creates a SymlinkDeletev20BadRequest with default headers values
func NewSymlinkDeletev20BadRequest() *SymlinkDeletev20BadRequest {
	return &SymlinkDeletev20BadRequest{}
}

/*SymlinkDeletev20BadRequest handles this case with default header values.

Invalid file system or path
*/
type SymlinkDeletev20BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *SymlinkDeletev20BadRequest) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/filesystems/{filesystemName}/filesets/{filesetName}/symlink/{path}][%d] symlinkDeletev20BadRequest  %+v", 400, o.Payload)
}

func (o *SymlinkDeletev20BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *SymlinkDeletev20BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSymlinkDeletev20InternalServerError creates a SymlinkDeletev20InternalServerError with default headers values
func NewSymlinkDeletev20InternalServerError() *SymlinkDeletev20InternalServerError {
	return &SymlinkDeletev20InternalServerError{}
}

/*SymlinkDeletev20InternalServerError handles this case with default header values.

An unexpected error occurred.
*/
type SymlinkDeletev20InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *SymlinkDeletev20InternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/filesystems/{filesystemName}/filesets/{filesetName}/symlink/{path}][%d] symlinkDeletev20InternalServerError  %+v", 500, o.Payload)
}

func (o *SymlinkDeletev20InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *SymlinkDeletev20InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
