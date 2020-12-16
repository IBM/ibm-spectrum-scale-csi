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

// QuotaDefaultsFilesetGetv2Reader is a Reader for the QuotaDefaultsFilesetGetv2 structure.
type QuotaDefaultsFilesetGetv2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *QuotaDefaultsFilesetGetv2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewQuotaDefaultsFilesetGetv2OK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewQuotaDefaultsFilesetGetv2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewQuotaDefaultsFilesetGetv2OK creates a QuotaDefaultsFilesetGetv2OK with default headers values
func NewQuotaDefaultsFilesetGetv2OK() *QuotaDefaultsFilesetGetv2OK {
	return &QuotaDefaultsFilesetGetv2OK{}
}

/*QuotaDefaultsFilesetGetv2OK handles this case with default header values.

successful operation
*/
type QuotaDefaultsFilesetGetv2OK struct {
	Payload *models.QuotaDefaultsInlineResponse200
}

func (o *QuotaDefaultsFilesetGetv2OK) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/filesystems/{filesystemName}/filesets/{filesetName}/quotadefaults][%d] quotaDefaultsFilesetGetv2OK  %+v", 200, o.Payload)
}

func (o *QuotaDefaultsFilesetGetv2OK) GetPayload() *models.QuotaDefaultsInlineResponse200 {
	return o.Payload
}

func (o *QuotaDefaultsFilesetGetv2OK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.QuotaDefaultsInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewQuotaDefaultsFilesetGetv2InternalServerError creates a QuotaDefaultsFilesetGetv2InternalServerError with default headers values
func NewQuotaDefaultsFilesetGetv2InternalServerError() *QuotaDefaultsFilesetGetv2InternalServerError {
	return &QuotaDefaultsFilesetGetv2InternalServerError{}
}

/*QuotaDefaultsFilesetGetv2InternalServerError handles this case with default header values.

Internal Server Error
*/
type QuotaDefaultsFilesetGetv2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *QuotaDefaultsFilesetGetv2InternalServerError) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/filesystems/{filesystemName}/filesets/{filesetName}/quotadefaults][%d] quotaDefaultsFilesetGetv2InternalServerError  %+v", 500, o.Payload)
}

func (o *QuotaDefaultsFilesetGetv2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *QuotaDefaultsFilesetGetv2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
