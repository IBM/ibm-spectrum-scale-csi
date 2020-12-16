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

// SmbSharesGetReader is a Reader for the SmbSharesGet structure.
type SmbSharesGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *SmbSharesGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewSmbSharesGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewSmbSharesGetBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewSmbSharesGetInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewSmbSharesGetOK creates a SmbSharesGetOK with default headers values
func NewSmbSharesGetOK() *SmbSharesGetOK {
	return &SmbSharesGetOK{}
}

/*SmbSharesGetOK handles this case with default header values.

successful operation
*/
type SmbSharesGetOK struct {
	Payload *models.SmbShareInlineResponse200
}

func (o *SmbSharesGetOK) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/smb/shares][%d] smbSharesGetOK  %+v", 200, o.Payload)
}

func (o *SmbSharesGetOK) GetPayload() *models.SmbShareInlineResponse200 {
	return o.Payload
}

func (o *SmbSharesGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.SmbShareInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSmbSharesGetBadRequest creates a SmbSharesGetBadRequest with default headers values
func NewSmbSharesGetBadRequest() *SmbSharesGetBadRequest {
	return &SmbSharesGetBadRequest{}
}

/*SmbSharesGetBadRequest handles this case with default header values.

SMB share not found
*/
type SmbSharesGetBadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *SmbSharesGetBadRequest) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/smb/shares][%d] smbSharesGetBadRequest  %+v", 400, o.Payload)
}

func (o *SmbSharesGetBadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *SmbSharesGetBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSmbSharesGetInternalServerError creates a SmbSharesGetInternalServerError with default headers values
func NewSmbSharesGetInternalServerError() *SmbSharesGetInternalServerError {
	return &SmbSharesGetInternalServerError{}
}

/*SmbSharesGetInternalServerError handles this case with default header values.

Internal Server Error
*/
type SmbSharesGetInternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *SmbSharesGetInternalServerError) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/smb/shares][%d] smbSharesGetInternalServerError  %+v", 500, o.Payload)
}

func (o *SmbSharesGetInternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *SmbSharesGetInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
