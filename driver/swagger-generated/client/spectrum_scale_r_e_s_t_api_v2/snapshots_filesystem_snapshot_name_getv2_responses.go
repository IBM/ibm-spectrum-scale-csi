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

// SnapshotsFilesystemSnapshotNameGetv2Reader is a Reader for the SnapshotsFilesystemSnapshotNameGetv2 structure.
type SnapshotsFilesystemSnapshotNameGetv2Reader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *SnapshotsFilesystemSnapshotNameGetv2Reader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewSnapshotsFilesystemSnapshotNameGetv2OK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewSnapshotsFilesystemSnapshotNameGetv2BadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewSnapshotsFilesystemSnapshotNameGetv2InternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewSnapshotsFilesystemSnapshotNameGetv2OK creates a SnapshotsFilesystemSnapshotNameGetv2OK with default headers values
func NewSnapshotsFilesystemSnapshotNameGetv2OK() *SnapshotsFilesystemSnapshotNameGetv2OK {
	return &SnapshotsFilesystemSnapshotNameGetv2OK{}
}

/*SnapshotsFilesystemSnapshotNameGetv2OK handles this case with default header values.

successful operation
*/
type SnapshotsFilesystemSnapshotNameGetv2OK struct {
	Payload *models.SnapshotInlineResponse200
}

func (o *SnapshotsFilesystemSnapshotNameGetv2OK) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/filesystems/{filesystemName}/snapshots/{snapshotName}][%d] snapshotsFilesystemSnapshotNameGetv2OK  %+v", 200, o.Payload)
}

func (o *SnapshotsFilesystemSnapshotNameGetv2OK) GetPayload() *models.SnapshotInlineResponse200 {
	return o.Payload
}

func (o *SnapshotsFilesystemSnapshotNameGetv2OK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.SnapshotInlineResponse200)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSnapshotsFilesystemSnapshotNameGetv2BadRequest creates a SnapshotsFilesystemSnapshotNameGetv2BadRequest with default headers values
func NewSnapshotsFilesystemSnapshotNameGetv2BadRequest() *SnapshotsFilesystemSnapshotNameGetv2BadRequest {
	return &SnapshotsFilesystemSnapshotNameGetv2BadRequest{}
}

/*SnapshotsFilesystemSnapshotNameGetv2BadRequest handles this case with default header values.

Invalid request
*/
type SnapshotsFilesystemSnapshotNameGetv2BadRequest struct {
	Payload *models.Http400BadRequest
}

func (o *SnapshotsFilesystemSnapshotNameGetv2BadRequest) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/filesystems/{filesystemName}/snapshots/{snapshotName}][%d] snapshotsFilesystemSnapshotNameGetv2BadRequest  %+v", 400, o.Payload)
}

func (o *SnapshotsFilesystemSnapshotNameGetv2BadRequest) GetPayload() *models.Http400BadRequest {
	return o.Payload
}

func (o *SnapshotsFilesystemSnapshotNameGetv2BadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http400BadRequest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewSnapshotsFilesystemSnapshotNameGetv2InternalServerError creates a SnapshotsFilesystemSnapshotNameGetv2InternalServerError with default headers values
func NewSnapshotsFilesystemSnapshotNameGetv2InternalServerError() *SnapshotsFilesystemSnapshotNameGetv2InternalServerError {
	return &SnapshotsFilesystemSnapshotNameGetv2InternalServerError{}
}

/*SnapshotsFilesystemSnapshotNameGetv2InternalServerError handles this case with default header values.

Internal Server Error
*/
type SnapshotsFilesystemSnapshotNameGetv2InternalServerError struct {
	Payload *models.Http500InternalServerError
}

func (o *SnapshotsFilesystemSnapshotNameGetv2InternalServerError) Error() string {
	return fmt.Sprintf("[GET /scalemgmt/v2/filesystems/{filesystemName}/snapshots/{snapshotName}][%d] snapshotsFilesystemSnapshotNameGetv2InternalServerError  %+v", 500, o.Payload)
}

func (o *SnapshotsFilesystemSnapshotNameGetv2InternalServerError) GetPayload() *models.Http500InternalServerError {
	return o.Payload
}

func (o *SnapshotsFilesystemSnapshotNameGetv2InternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Http500InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
