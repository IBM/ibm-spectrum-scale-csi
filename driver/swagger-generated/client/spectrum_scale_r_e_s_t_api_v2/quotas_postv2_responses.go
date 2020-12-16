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

// QuotasPostv2Reader is a Reader for the QuotasPostv2 structure.
type QuotasPostv2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *QuotasPostv2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewQuotasPostv2Accepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewQuotasPostv2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewQuotasPostv2Accepted creates a QuotasPostv2Accepted with default headers values
func NewQuotasPostv2Accepted() *QuotasPostv2Accepted {
	return &QuotasPostv2Accepted{}
}

/*QuotasPostv2Accepted handles this case with default header values.

successful operation
*/
type QuotasPostv2Accepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *QuotasPostv2Accepted) Error() string {
	return fmt.Sprintf("[POST /scalemgmt/v2/filesystems/{filesystemName}/quotas][%d] quotasPostv2Accepted  %+v", 202, o.Payload)
}

func (o *QuotasPostv2Accepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *QuotasPostv2Accepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewQuotasPostv2InternalServerError creates a QuotasPostv2InternalServerError with default headers values
func NewQuotasPostv2InternalServerError() *QuotasPostv2InternalServerError {
	return &QuotasPostv2InternalServerError{}
}

/*QuotasPostv2InternalServerError handles this case with default header values.

Internal Server Error
*/
type QuotasPostv2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *QuotasPostv2InternalServerError) Error() string {
	return fmt.Sprintf("[POST /scalemgmt/v2/filesystems/{filesystemName}/quotas][%d] quotasPostv2InternalServerError  %+v", 500, o.Payload)
}

func (o *QuotasPostv2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *QuotasPostv2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
