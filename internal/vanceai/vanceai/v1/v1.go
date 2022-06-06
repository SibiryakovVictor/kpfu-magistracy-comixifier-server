package v1

//go:generate mockgen -destination ./mock/client.go -package mock --build_flags=--mod=mod integration-vanceai/internal/http/vanceai/v1 Client
//go:generate mockgen -destination ./mock/decoder.go -package mock --build_flags=--mod=mod integration-vanceai/internal/json/vanceai/v1 ResponseDecoder
//go:generate mockgen -destination ./mock/encoder.go -package mock --build_flags=--mod=mod integration-vanceai/internal/json/vanceai/v1 JConfigEncoder

import (
	"comixifier/internal/vanceai/filesystem"
	v1http "comixifier/internal/vanceai/http/vanceai/v1"
	v1json "comixifier/internal/vanceai/json/vanceai/v1"
	"comixifier/internal/vanceai/json/vanceai/v1/jconfig"
	"comixifier/internal/vanceai/vanceai/v1/errors"
	"comixifier/internal/vanceai/vanceai/v1/image"
	"fmt"
	"io"
)

type VanceAI interface {
	Upload(r *UploadRequest) (*UploadResponse, error)
	Transform(r *TransformRequest) (*TransformResponse, error)
	Progress(r *ProgressRequest) (*ProgressResponse, error)
	Download(r *DownloadRequest) (*DownloadResponse, error)
}

type vanceAI struct {
	client         v1http.Client
	respDecoder    v1json.ResponseDecoder
	jConfigEncoder v1json.JConfigEncoder
}

func NewVanceAI(
	client v1http.Client,
	respDecoder v1json.ResponseDecoder,
	jConfigEncoder v1json.JConfigEncoder,
) *vanceAI {
	return &vanceAI{
		client:         client,
		respDecoder:    respDecoder,
		jConfigEncoder: jConfigEncoder,
	}
}

func (v *vanceAI) Upload(r *UploadRequest) (*UploadResponse, error) {
	clientReq := v1http.NewUploadRequest(r.job, r.file)

	respReader, err := v.client.SendUploadRequest(clientReq)
	if err != nil {
		return nil, fmt.Errorf("send Upload request: %w", err)
	}
	defer respReader.Close()

	resp, err := io.ReadAll(respReader)
	if err != nil {
		return nil, fmt.Errorf("read response content: %w", err)
	}

	decoderRes, err := v.respDecoder.ToUploadResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("decode upload response from json: %w", err)
	}

	if decoderRes.IsError() {
		errCode, err := errors.MapCode(decoderRes.Code())
		if err != nil {
			return nil, fmt.Errorf("map error code from response: %w", err)
		}

		return nil, errors.NewApiError(errCode, decoderRes.Msg())
	}

	methodResp := NewUploadResponse(decoderRes.Uid())
	return methodResp, nil
}

func (v *vanceAI) Transform(r *TransformRequest) (*TransformResponse, error) {
	var jConfig io.Reader
	var err error
	if len(r.processors) == 1 {
		job := jconfig.NewSingleJob(r.processors[0].Map())
		jConfig, err = v.jConfigEncoder.DoSingleJob(job)
		if err != nil {
			return nil, fmt.Errorf("encode single job: %w", err)
		}
	} else {
		var features []jconfig.Feature
		for _, processor := range r.processors {
			features = append(features, processor.Map())
		}
		workflow := jconfig.NewWorkflow(features)
		jConfig, err = v.jConfigEncoder.DoWorkflow(workflow)
		if err != nil {
			return nil, fmt.Errorf("encode workflow: %w", err)
		}
	}

	httpResp, err := v.client.SendTransformRequest(v1http.NewTransformRequest(r.uid, jConfig))
	defer httpResp.Close()
	if err != nil {
		return nil, fmt.Errorf("send transform request: %w", err)
	}

	methodHttpResp, err := v.respDecoder.ToTransformResponse(httpResp)
	if err != nil {
		return nil, fmt.Errorf("decode transform response: %w", err)
	}
	if methodHttpResp.IsError() {
		return nil, fmt.Errorf("transform response error: %s", methodHttpResp.Msg())
	}

	jobId := JobId(methodHttpResp.JobId())
	var jobStatus JobStatus
	switch methodHttpResp.JobStatus() {
	case "process":
		jobStatus = JobStatusProcess
	}

	return NewTransformResponse(jobId, jobStatus), nil
}

func (v *vanceAI) Progress(r *ProgressRequest) (*ProgressResponse, error) {
	if r.id == "" {
		return nil, fmt.Errorf("method request: empty jobId")
	}

	httpReq := v1http.NewProgressRequest(string(r.id))

	respReader, err := v.client.SendProgressRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer respReader.Close()

	methodHttpResp, err := v.respDecoder.ToProgressResponse(respReader)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// TODO: case where methodHttpResp.IsError() == true

	var jobStatus JobStatus
	switch methodHttpResp.JobStatus() {
	case "process":
		jobStatus = JobStatusProcess
	case "waiting":
		jobStatus = JobStatusWaiting
	case "finish":
		jobStatus = JobStatusFinish
	case "fatal":
		jobStatus = JobStatusFatal
	default:
		return nil, fmt.Errorf("unexpected job status in response: %s", methodHttpResp.JobStatus())
	}

	return NewProgressResponse(jobStatus, methodHttpResp.Filesize()), nil
}

func (v *vanceAI) Download(r *DownloadRequest) (*DownloadResponse, error) {
	httpReq := v1http.NewDownloadRequest(string(r.Id()))

	resp, err := v.client.SendDownloadRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send http request: %w", err)
	}

	switch resp.ContentType() {
	case v1http.ImgContentType:
		return NewDownloadResponse(resp.Content()), nil
	case v1http.JsonContentType:
		errResp, err := v.respDecoder.ToErrorResponse(resp.Content())
		if err != nil {
			return nil, fmt.Errorf("decode error response: %w", err)
		}

		errCode, err := errors.MapCode(errResp.Code())
		if err != nil {
			return nil, fmt.Errorf("map error code: %w", err)
		}

		return nil, errors.NewApiError(errCode, errResp.Msg())
	default:
		return nil, fmt.Errorf("unexpected type of response: %d", resp.ContentType())
	}
}

type UploadRequest struct {
	job  string
	file filesystem.File
}

func NewUploadRequest(job string, file filesystem.File) *UploadRequest {
	return &UploadRequest{
		job:  job,
		file: file,
	}
}

func (r *UploadRequest) File() filesystem.File {
	return r.file
}

func (r *UploadRequest) Job() string {
	return r.job
}

type UploadResponse struct {
	uid string
}

func NewUploadResponse(uid string) *UploadResponse {
	return &UploadResponse{
		uid: uid,
	}
}

func (r *UploadResponse) Uid() string {
	return r.uid
}

type TransformRequest struct {
	uid        string
	processors []image.Processor
}

func NewTransformRequest(uid string, processors []image.Processor) *TransformRequest {
	return &TransformRequest{
		uid:        uid,
		processors: processors,
	}
}

func (r *TransformRequest) Uid() string {
	return r.uid
}

func (r *TransformRequest) Processors() []image.Processor {
	return r.processors
}

type TransformResponse struct {
	id     JobId
	status JobStatus
}

func NewTransformResponse(id JobId, status JobStatus) *TransformResponse {
	return &TransformResponse{id: id, status: status}
}

type ProgressRequest struct {
	id JobId
}

func NewProgressRequest(id JobId) *ProgressRequest {
	return &ProgressRequest{id: id}
}

func (r *ProgressRequest) Id() JobId {
	return r.id
}

type ProgressResponse struct {
	status   JobStatus
	filesize uint32
}

func NewProgressResponse(status JobStatus, filesize uint32) *ProgressResponse {
	return &ProgressResponse{status: status, filesize: filesize}
}

func (r *ProgressResponse) Status() JobStatus {
	return r.status
}

func (r *ProgressResponse) Filesize() uint32 {
	return r.filesize
}

type DownloadRequest struct {
	id JobId
}

func NewDownloadRequest(id JobId) *DownloadRequest {
	return &DownloadRequest{id: id}
}

func (r *DownloadRequest) Id() JobId {
	return r.id
}

type DownloadResponse struct {
	imgContent io.ReadCloser
}

func NewDownloadResponse(imgContent io.ReadCloser) *DownloadResponse {
	return &DownloadResponse{imgContent: imgContent}
}

func (r *DownloadResponse) ImgContent() io.ReadCloser {
	return r.imgContent
}

type JobId string
type JobStatus uint8

func (s *JobStatus) String() string {
	switch *s {
	case JobStatusProcess:
		return "process"
	case JobStatusWaiting:
		return "waiting"
	case JobStatusFatal:
		return "fatal"
	case JobStatusFinish:
		return "finish"
	default:
		return "unknown"
	}
}

var (
	JobStatusProcess = JobStatus(0)
	JobStatusWaiting = JobStatus(1)
	JobStatusFinish  = JobStatus(2)
	JobStatusFatal   = JobStatus(3)
)
