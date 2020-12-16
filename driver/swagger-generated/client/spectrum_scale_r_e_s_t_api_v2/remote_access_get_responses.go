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

// RemoteAccessGetReader is a Reader for the RemoteAccessGet structure.
type RemoteAccessGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *RemoteAccessGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewRemoteAccessGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewRemoteAccessGetInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewRemoteAccessGetOK creates a RemoteAccessGetOK with default headers values
func NewRemoteAccessGetOK() *RemoteAccessGetOK {
	return &RemoteAccessGetOK{}
}

/*RemoteAccessGetOK handles this case with default header values.

successful operation
*/
type RemoteAccessGetOK struct {
	Payload *models.RemoteAccessInlineResponse200
}

func (o *RemoteAccessGetOK) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/access][%d] remoteAccessGetOK  %+v", 200, o.Payload)
}

func (o *RemoteAccessGetOK) GetPayload() *models.RemoteAccessInlineResponse200 {
	return o.Payload
}

func (o *RemoteAccessGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RemoteAccessInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewRemoteAccessGetInternalServerError creates a RemoteAccessGetInternalServerError with default headers values
func NewRemoteAccessGetInternalServerError() *RemoteAccessGetInternalServerError {
	return &RemoteAccessGetInternalServerError{}
}

/*RemoteAccessGetInternalServerError handles this case with default header values.

Internal Server Error
*/
type RemoteAccessGetInternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *RemoteAccessGetInternalServerError) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/access][%d] remoteAccessGetInternalServerError  %+v", 500, o.Payload)
}

func (o *RemoteAccessGetInternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *RemoteAccessGetInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
