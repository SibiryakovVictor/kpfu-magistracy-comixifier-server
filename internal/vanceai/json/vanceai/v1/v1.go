package v1

import (
	"comixifier/internal/vanceai/json/vanceai/v1/jconfig"
	"io"
)

type ResponseDecoder interface {
	ToUploadResponse(data []byte) (*UploadResponse, error)
	ToTransformResponse(data io.Reader) (*TransformResponse, error)
	ToProgressResponse(data io.Reader) (*ProgressResponse, error)
	ToErrorResponse(data io.Reader) (*Response, error)
}

type Response struct {
	code int
	msg  string
}

func NewResponse(code int, msg string) *Response {
	return &Response{
		code: code,
		msg:  msg,
	}
}

func (r *Response) Code() int {
	return r.code
}

func (r *Response) Msg() string {
	return r.msg
}

func (r *Response) IsError() bool {
	return IsErrorCode(r.code)
}

type UploadResponse struct {
	Response
	uid string
}

func NewUploadResponse(code int, uid string) *UploadResponse {
	return &UploadResponse{
		Response: Response{
			code: code,
		},
		uid: uid,
	}
}

func NewUploadErrorResponse(code int, msg string) *UploadResponse {
	return &UploadResponse{
		Response: Response{
			code: code,
			msg:  msg,
		},
	}
}

func (r *UploadResponse) Uid() string {
	return r.uid
}

type TransformResponse struct {
	Response
	id     string
	status string
}

func (r *TransformResponse) JobId() string {
	return r.id
}

func (r *TransformResponse) JobStatus() string {
	return r.status
}

func NewTransformResponse(code int, jobId string, status string) *TransformResponse {
	return &TransformResponse{
		Response: Response{
			code: code,
		},
		id:     jobId,
		status: status,
	}
}

func NewTransformErrorResponse(code int, msg string) *TransformResponse {
	return &TransformResponse{
		Response: Response{
			code: code,
			msg:  msg,
		},
	}
}

type ProgressResponse struct {
	Response
	jobStatus string
	filesize  uint32
}

func (r *ProgressResponse) JobStatus() string {
	return r.jobStatus
}

func (r *ProgressResponse) Filesize() uint32 {
	return r.filesize
}

func NewProgressResponse(code int, jobStatus string, filesize uint32) *ProgressResponse {
	return &ProgressResponse{
		Response: Response{
			code: code,
		},
		jobStatus: jobStatus,
		filesize:  filesize,
	}
}

func NewProgressErrorResponse(code int, msg string) *ProgressResponse {
	return &ProgressResponse{
		Response: Response{
			code: code,
			msg:  msg,
		},
	}
}

type JConfigEncoder interface {
	DoSingleJob(j *jconfig.SingleJob) (io.Reader, error)
	DoWorkflow(j *jconfig.Workflow) (io.Reader, error)
}

func IsErrorCode(code int) bool {
	return code >= 10001
}
