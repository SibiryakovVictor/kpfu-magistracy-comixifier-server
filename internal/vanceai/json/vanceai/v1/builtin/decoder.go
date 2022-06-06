package builtin

import (
	v1 "comixifier/internal/vanceai/json/vanceai/v1"
	"comixifier/internal/vanceai/json/vanceai/v1/builtin/schema"
	"encoding/json"
	"fmt"
	"io"
)

type ResponseDecoder struct {
}

func NewResponseDecoder() *ResponseDecoder {
	return &ResponseDecoder{}
}

func (d *ResponseDecoder) ToUploadResponse(data []byte) (*v1.UploadResponse, error) {
	errorRespJson := &schema.ErrorResponse{}
	err := json.Unmarshal(data, errorRespJson)
	if err != nil {
		return nil, fmt.Errorf("try decode response as error: %w", err)
	}

	if v1.IsErrorCode(errorRespJson.Code) {
		msg, err := errorRespJson.MsgString()
		if err != nil {
			return nil, fmt.Errorf("msg of error: %w", err)
		}

		return v1.NewUploadErrorResponse(errorRespJson.Code, msg), nil
	}

	methodRespJson := &schema.UploadImageResponse{}
	err = json.Unmarshal(data, methodRespJson)
	if err != nil {
		return nil, fmt.Errorf("try decode response: %w", err)
	}

	return v1.NewUploadResponse(methodRespJson.Code, methodRespJson.Data.Uid), nil
}

func (d *ResponseDecoder) ToTransformResponse(data io.Reader) (*v1.TransformResponse, error) {
	dataBytes, _ := io.ReadAll(data)

	errorResp := &schema.ErrorResponse{}
	err := json.Unmarshal(dataBytes, errorResp)
	if err != nil {
		return nil, fmt.Errorf("decode to ErrorResponse: %w", err)
	}

	if v1.IsErrorCode(errorResp.Code) {
		msg, _ := errorResp.MsgString()
		return v1.NewTransformErrorResponse(errorResp.Code, msg), nil
	}

	methodResp := &schema.TransformResponse{}
	json.Unmarshal(dataBytes, methodResp)

	return v1.NewTransformResponse(methodResp.Code, methodResp.Data.TransId, methodResp.Data.Status), nil
}

func (d *ResponseDecoder) ToProgressResponse(data io.Reader) (*v1.ProgressResponse, error) {
	dataBytes, _ := io.ReadAll(data)

	errorResp := &schema.ErrorResponse{}
	err := json.Unmarshal(dataBytes, errorResp)
	if err != nil {
		return nil, fmt.Errorf("decode to ErrorResponse: %w", err)
	}

	if v1.IsErrorCode(errorResp.Code) {
		msg, _ := errorResp.MsgString()
		return v1.NewProgressErrorResponse(errorResp.Code, msg), nil
	}

	methodResp := &schema.ProgressResponse{}
	json.Unmarshal(dataBytes, methodResp)

	return v1.NewProgressResponse(methodResp.Code, methodResp.Data.Status, methodResp.Data.Filesize), nil
}

func (d *ResponseDecoder) ToErrorResponse(data io.Reader) (*v1.Response, error) {
	errorResp := &schema.ErrorResponse{}
	err := json.NewDecoder(data).Decode(errorResp)
	if err != nil {
		return nil, fmt.Errorf("decode data to ErrorResponse: %w", err)
	}

	msg, err := errorResp.MsgString()
	if err != nil {
		return nil, fmt.Errorf("get error msg: %w", err)
	}

	return v1.NewResponse(errorResp.Code, msg), nil
}
