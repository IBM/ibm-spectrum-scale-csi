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

// JobsDeletev2Reader is a Reader for the JobsDeletev2 structure.
type JobsDeletev2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *JobsDeletev2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewJobsDeletev2OK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewJobsDeletev2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewJobsDeletev2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewJobsDeletev2OK creates a JobsDeletev2OK with default headers values
func NewJobsDeletev2OK() *JobsDeletev2OK {
	return &JobsDeletev2OK{}
}

/*JobsDeletev2OK handles this case with default header values.

successful operation
*/
type JobsDeletev2OK struct {
	Payload *models.JobInlineResponse200
}

func (o *JobsDeletev2OK) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/jobs/{jobId}][%d] jobsDeletev2OK  %+v", 200, o.Payload)
}

func (o *JobsDeletev2OK) GetPayload() *models.JobInlineResponse200 {
	return o.Payload
}

func (o *JobsDeletev2OK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.JobInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewJobsDeletev2BadRequest creates a JobsDeletev2BadRequest with default headers values
func NewJobsDeletev2BadRequest() *JobsDeletev2BadRequest {
	return &JobsDeletev2BadRequest{}
}

/*JobsDeletev2BadRequest handles this case with default header values.

Invalid request
*/
type JobsDeletev2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *JobsDeletev2BadRequest) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/jobs/{jobId}][%d] jobsDeletev2BadRequest  %+v", 400, o.Payload)
}

func (o *JobsDeletev2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *JobsDeletev2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewJobsDeletev2InternalServerError creates a JobsDeletev2InternalServerError with default headers values
func NewJobsDeletev2InternalServerError() *JobsDeletev2InternalServerError {
	return &JobsDeletev2InternalServerError{}
}

/*JobsDeletev2InternalServerError handles this case with default header values.

Internal Server Error
*/
type JobsDeletev2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *JobsDeletev2InternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/jobs/{jobId}][%d] jobsDeletev2InternalServerError  %+v", 500, o.Payload)
}

func (o *JobsDeletev2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *JobsDeletev2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
