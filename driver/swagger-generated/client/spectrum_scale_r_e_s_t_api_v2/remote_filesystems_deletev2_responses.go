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

// RemoteFilesystemsDeletev2Reader is a Reader for the RemoteFilesystemsDeletev2 structure.
type RemoteFilesystemsDeletev2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RemoteFilesystemsDeletev2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewRemoteFilesystemsDeletev2Accepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewRemoteFilesystemsDeletev2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewRemoteFilesystemsDeletev2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewRemoteFilesystemsDeletev2Accepted creates a RemoteFilesystemsDeletev2Accepted with default headers values
func NewRemoteFilesystemsDeletev2Accepted() *RemoteFilesystemsDeletev2Accepted {
	return &RemoteFilesystemsDeletev2Accepted{}
}

/*RemoteFilesystemsDeletev2Accepted handles this case with default header values.

successful operation
*/
type RemoteFilesystemsDeletev2Accepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *RemoteFilesystemsDeletev2Accepted) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/remotemount/remotefilesystems/{remoteFilesystem}][%d] remoteFilesystemsDeletev2Accepted  %+v", 202, o.Payload)
}

func (o *RemoteFilesystemsDeletev2Accepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *RemoteFilesystemsDeletev2Accepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoteFilesystemsDeletev2BadRequest creates a RemoteFilesystemsDeletev2BadRequest with default headers values
func NewRemoteFilesystemsDeletev2BadRequest() *RemoteFilesystemsDeletev2BadRequest {
	return &RemoteFilesystemsDeletev2BadRequest{}
}

/*RemoteFilesystemsDeletev2BadRequest handles this case with default header values.

Invalid request
*/
type RemoteFilesystemsDeletev2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *RemoteFilesystemsDeletev2BadRequest) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/remotemount/remotefilesystems/{remoteFilesystem}][%d] remoteFilesystemsDeletev2BadRequest  %+v", 400, o.Payload)
}

func (o *RemoteFilesystemsDeletev2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *RemoteFilesystemsDeletev2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoteFilesystemsDeletev2InternalServerError creates a RemoteFilesystemsDeletev2InternalServerError with default headers values
func NewRemoteFilesystemsDeletev2InternalServerError() *RemoteFilesystemsDeletev2InternalServerError {
	return &RemoteFilesystemsDeletev2InternalServerError{}
}

/*RemoteFilesystemsDeletev2InternalServerError handles this case with default header values.

Internal Server Error
*/
type RemoteFilesystemsDeletev2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *RemoteFilesystemsDeletev2InternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/remotemount/remotefilesystems/{remoteFilesystem}][%d] remoteFilesystemsDeletev2InternalServerError  %+v", 500, o.Payload)
}

func (o *RemoteFilesystemsDeletev2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *RemoteFilesystemsDeletev2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
