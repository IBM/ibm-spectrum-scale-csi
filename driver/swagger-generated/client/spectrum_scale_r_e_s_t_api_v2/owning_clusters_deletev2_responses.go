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

// OwningClustersDeletev2Reader is a Reader for the OwningClustersDeletev2 structure.
type OwningClustersDeletev2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *OwningClustersDeletev2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 202:
		result := NewOwningClustersDeletev2Accepted()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewOwningClustersDeletev2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewOwningClustersDeletev2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewOwningClustersDeletev2Accepted creates a OwningClustersDeletev2Accepted with default headers values
func NewOwningClustersDeletev2Accepted() *OwningClustersDeletev2Accepted {
	return &OwningClustersDeletev2Accepted{}
}

/*OwningClustersDeletev2Accepted handles this case with default header values.

successful operation
*/
type OwningClustersDeletev2Accepted struct {
	Payload *models.AsyncRequestResponse
}

func (o *OwningClustersDeletev2Accepted) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/remotemount/owningclusters/{owningCluster}][%d] owningClustersDeletev2Accepted  %+v", 202, o.Payload)
}

func (o *OwningClustersDeletev2Accepted) GetPayload() *models.AsyncRequestResponse {
	return o.Payload
}

func (o *OwningClustersDeletev2Accepted) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AsyncRequestResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewOwningClustersDeletev2BadRequest creates a OwningClustersDeletev2BadRequest with default headers values
func NewOwningClustersDeletev2BadRequest() *OwningClustersDeletev2BadRequest {
	return &OwningClustersDeletev2BadRequest{}
}

/*OwningClustersDeletev2BadRequest handles this case with default header values.

Invalid request
*/
type OwningClustersDeletev2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *OwningClustersDeletev2BadRequest) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/remotemount/owningclusters/{owningCluster}][%d] owningClustersDeletev2BadRequest  %+v", 400, o.Payload)
}

func (o *OwningClustersDeletev2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *OwningClustersDeletev2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewOwningClustersDeletev2InternalServerError creates a OwningClustersDeletev2InternalServerError with default headers values
func NewOwningClustersDeletev2InternalServerError() *OwningClustersDeletev2InternalServerError {
	return &OwningClustersDeletev2InternalServerError{}
}

/*OwningClustersDeletev2InternalServerError handles this case with default header values.

Internal Server Error
*/
type OwningClustersDeletev2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *OwningClustersDeletev2InternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /scalemgmt/v2/remotemount/owningclusters/{owningCluster}][%d] owningClustersDeletev2InternalServerError  %+v", 500, o.Payload)
}

func (o *OwningClustersDeletev2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *OwningClustersDeletev2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
