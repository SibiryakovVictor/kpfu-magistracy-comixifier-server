package v1

import (
	"comixifier/internal/vanceai/filesystem"
	"io"
)

type Client interface {
	SendUploadRequest(r *UploadRequest) (io.ReadCloser, error)
	SendTransformRequest(r *TransformRequest) (io.ReadCloser, error)
	SendProgressRequest(r *ProgressRequest) (io.ReadCloser, error)
	SendDownloadRequest(r *DownloadRequest) (*Response, error)
}

type Response struct {
	content     io.ReadCloser
	contentType ContentType
}

func NewResponse(content io.ReadCloser, contentType ContentType) *Response {
	return &Response{content: content, contentType: contentType}
}

func (r *Response) Content() io.ReadCloser {
	return r.content
}

func (r *Response) ContentType() ContentType {
	return r.contentType
}

type ContentType uint8

const (
	UnknownContentType ContentType = iota
	JsonContentType
	ImgContentType
)

type UploadRequest struct {
	job   string
	image filesystem.File
}

func NewUploadRequest(job string, image filesystem.File) *UploadRequest {
	return &UploadRequest{job: job, image: image}
}

func (r *UploadRequest) Job() string {
	return r.job
}

func (r *UploadRequest) Image() filesystem.File {
	return r.image
}

type TransformRequest struct {
	uid     string
	jConfig io.Reader
}

func NewTransformRequest(uid string, jConfig io.Reader) *TransformRequest {
	return &TransformRequest{
		uid:     uid,
		jConfig: jConfig,
	}
}

func (r *TransformRequest) Uid() string {
	return r.uid
}

func (r *TransformRequest) JConfig() io.Reader {
	return r.jConfig
}

type ProgressRequest struct {
	jobId string
}

func NewProgressRequest(jobId string) *ProgressRequest {
	return &ProgressRequest{jobId: jobId}
}

func (r *ProgressRequest) JobId() string {
	return r.jobId
}

type DownloadRequest struct {
	jobId string
}

func NewDownloadRequest(jobId string) *DownloadRequest {
	return &DownloadRequest{jobId: jobId}
}

func (r *DownloadRequest) JobId() string {
	return r.jobId
}
