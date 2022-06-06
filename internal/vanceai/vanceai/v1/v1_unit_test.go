package v1

import (
	v1 "comixifier/internal/vanceai/http/vanceai/v1"
	v12 "comixifier/internal/vanceai/json/vanceai/v1"
	"comixifier/internal/vanceai/json/vanceai/v1/jconfig"
	"comixifier/internal/vanceai/vanceai/v1/errors"
	"comixifier/internal/vanceai/vanceai/v1/image"
	"comixifier/internal/vanceai/vanceai/v1/mock"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type vanceAISuite struct {
	suite.Suite

	api *vanceAI

	ctrl           *gomock.Controller
	decoder        *mock.MockResponseDecoder
	client         *mock.MockClient
	jConfigEncoder *mock.MockJConfigEncoder
}

type testReadCloser struct {
	strings.Reader
}

func newTestReadCloser(content string) *testReadCloser {
	return &testReadCloser{Reader: *strings.NewReader(content)}
}

func (rc *testReadCloser) Close() error {
	return nil
}

func (s *vanceAISuite) TestTransform_Unit() {
	type testCase struct {
		name        string
		prepareArgs func() *TransformRequest
		wantOut     *TransformResponse
		wantErr     error
	}
	tests := []testCase{
		{
			name: "correct single job",
			prepareArgs: func() *TransformRequest {
				r := &TransformRequest{
					uid:        "fvfdv1231csdc2341231csad",
					processors: []image.Processor{image.NewCartoonizer()},
				}

				jConfigReader := strings.NewReader("jconfig content")
				s.jConfigEncoder.EXPECT().
					DoSingleJob(jconfig.NewSingleJob(r.processors[0].Map())).
					Return(jConfigReader, nil)

				httpReq := v1.NewTransformRequest(r.uid, jConfigReader)
				respReader := newTestReadCloser("response content")
				s.client.EXPECT().
					SendTransformRequest(httpReq).
					Return(respReader, nil)

				s.decoder.EXPECT().
					ToTransformResponse(respReader).
					Return(v12.NewTransformResponse(
						200,
						"6de4b562d1a01c3d2520608eae929646",
						"process",
					), nil)

				return r
			},
			wantOut: &TransformResponse{
				id:     "6de4b562d1a01c3d2520608eae929646",
				status: JobStatusProcess,
			},
			wantErr: nil,
		},
		{
			name: "can't encode jconfig",
			prepareArgs: func() *TransformRequest {
				r := &TransformRequest{
					uid:        "fvfdv1231csdc2341231csad",
					processors: []image.Processor{image.NewCartoonizer()},
				}

				s.jConfigEncoder.EXPECT().
					DoSingleJob(jconfig.NewSingleJob(r.processors[0].Map())).
					Return(nil, fmt.Errorf("some_error"))

				return r
			},
			wantOut: nil,
			wantErr: fmt.Errorf("encode single job: some_error"),
		},
		{
			name: "correct workflow",
			prepareArgs: func() *TransformRequest {
				r := &TransformRequest{
					uid:        "fvfdv1231csdc2341231csad",
					processors: []image.Processor{image.NewCartoonizer(), image.NewCartoonizer()},
				}

				jConfigReader := strings.NewReader("jconfig content")
				s.jConfigEncoder.EXPECT().
					DoWorkflow(jconfig.NewWorkflow([]jconfig.Feature{r.processors[0].Map(), r.processors[1].Map()})).
					Return(jConfigReader, nil)

				httpReq := v1.NewTransformRequest(r.uid, jConfigReader)
				respReader := newTestReadCloser("response content")
				s.client.EXPECT().
					SendTransformRequest(httpReq).
					Return(respReader, nil)

				s.decoder.EXPECT().
					ToTransformResponse(respReader).
					Return(v12.NewTransformResponse(
						200,
						"6de4b562d1a01c3d2520608eae929646",
						"process",
					), nil)

				return r
			},
			wantOut: &TransformResponse{
				id:     "6de4b562d1a01c3d2520608eae929646",
				status: JobStatusProcess,
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			r := test.prepareArgs()

			out, err := s.api.Transform(r)
			if test.wantErr != nil {
				s.Nil(out)
				s.EqualError(err, test.wantErr.Error())
			} else {
				s.Nil(err)
				s.Equal(test.wantOut, out)
			}
		})
	}
}

func (s *vanceAISuite) TestProgress_Unit() {
	type testCase struct {
		name        string
		prepareArgs func() *ProgressRequest
		wantOut     *ProgressResponse
		wantErr     error
	}
	tests := []testCase{
		{
			name: "correct with status = finish",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				httpResp := newTestReadCloser("response content")
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(httpResp, nil)

				s.decoder.EXPECT().
					ToProgressResponse(httpResp).
					Return(v12.NewProgressResponse(200, "finish", uint32(12345678)), nil)

				return r
			},
			wantOut: NewProgressResponse(JobStatusFinish, uint32(12345678)),
			wantErr: nil,
		},
		{
			name: "correct with status = waiting",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				httpResp := newTestReadCloser("response content")
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(httpResp, nil)

				s.decoder.EXPECT().
					ToProgressResponse(httpResp).
					Return(v12.NewProgressResponse(200, "waiting", uint32(12345678)), nil)

				return r
			},
			wantOut: NewProgressResponse(JobStatusWaiting, uint32(12345678)),
			wantErr: nil,
		},
		{
			name: "correct with status = fatal",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				httpResp := newTestReadCloser("response content")
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(httpResp, nil)

				s.decoder.EXPECT().
					ToProgressResponse(httpResp).
					Return(v12.NewProgressResponse(200, "fatal", uint32(12345678)), nil)

				return r
			},
			wantOut: NewProgressResponse(JobStatusFatal, uint32(12345678)),
			wantErr: nil,
		},
		{
			name: "correct with status = process",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				httpResp := newTestReadCloser("response content")
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(httpResp, nil)

				s.decoder.EXPECT().
					ToProgressResponse(httpResp).
					Return(v12.NewProgressResponse(200, "process", uint32(12345678)), nil)

				return r
			},
			wantOut: NewProgressResponse(JobStatusProcess, uint32(12345678)),
			wantErr: nil,
		},
		{
			name: "unexpected job status in response",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				httpResp := newTestReadCloser("response content")
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(httpResp, nil)

				s.decoder.EXPECT().
					ToProgressResponse(httpResp).
					Return(v12.NewProgressResponse(200, "random_job_status", uint32(12345678)), nil)

				return r
			},
			wantOut: nil,
			wantErr: fmt.Errorf("unexpected job status in response: %s", "random_job_status"),
		},
		{
			name: "send request",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(nil, fmt.Errorf("some error"))

				return r
			},
			wantOut: nil,
			wantErr: fmt.Errorf("send request: some error"),
		},
		{
			name: "decode response",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("some12job34id56")

				httpReq := v1.NewProgressRequest(string(r.Id()))
				httpResp := newTestReadCloser("response content")
				s.client.EXPECT().
					SendProgressRequest(httpReq).
					Return(httpResp, nil)

				s.decoder.EXPECT().
					ToProgressResponse(httpResp).
					Return(nil, fmt.Errorf("some error"))

				return r
			},
			wantOut: nil,
			wantErr: fmt.Errorf("decode response: some error"),
		},
		{
			name: "empty jobId in method request",
			prepareArgs: func() *ProgressRequest {
				r := NewProgressRequest("")

				return r
			},
			wantOut: nil,
			wantErr: fmt.Errorf("method request: empty jobId"),
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			r := test.prepareArgs()

			out, err := s.api.Progress(r)
			if test.wantErr != nil {
				s.Nil(out)
				s.EqualError(err, test.wantErr.Error())
			} else {
				s.Nil(err)
				s.Equal(test.wantOut, out)
			}
		})
	}
}

func (s *vanceAISuite) TestDownload_Unit() {
	type testCase struct {
		name        string
		prepareArgs func() (*DownloadRequest, *DownloadResponse)
		wantErr     error
	}
	tests := []testCase{
		{
			name: "correct when image",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				contentReader := newTestReadCloser("some_content")
				resp := v1.NewResponse(contentReader, v1.ImgContentType)
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(resp, nil)

				return methodReq, NewDownloadResponse(contentReader)
			},
			wantErr: nil,
		},
		{
			name: "correct when json with error",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				contentReader := newTestReadCloser("some_content")
				resp := v1.NewResponse(contentReader, v1.JsonContentType)
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(resp, nil)

				errorResp := v12.NewResponse(int(errors.CodeFileNotAvailable), "some_error_message")
				s.decoder.EXPECT().
					ToErrorResponse(contentReader).
					Return(errorResp, nil)

				return methodReq, nil
			},
			wantErr: errors.NewApiError(errors.CodeFileNotAvailable, "some_error_message"),
		},
		{
			name: "can't send request",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(nil, fmt.Errorf("some_error"))

				return methodReq, nil
			},
			wantErr: fmt.Errorf("send http request: some_error"),
		},
		{
			name: "unknown type of response",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				resp := v1.NewResponse(nil, v1.ContentType(123))
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(resp, nil)

				return methodReq, nil
			},
			wantErr: fmt.Errorf("unexpected type of response: %d", v1.ContentType(123)),
		},
		{
			name: "unknown type of response",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				resp := v1.NewResponse(nil, v1.ContentType(123))
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(resp, nil)

				return methodReq, nil
			},
			wantErr: fmt.Errorf("unexpected type of response: %d", v1.ContentType(123)),
		},
		{
			name: "decoder error",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				contentReader := newTestReadCloser("some_content")
				resp := v1.NewResponse(contentReader, v1.JsonContentType)
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(resp, nil)

				s.decoder.EXPECT().
					ToErrorResponse(contentReader).
					Return(nil, fmt.Errorf("some_error"))

				return methodReq, nil
			},
			wantErr: fmt.Errorf("decode error response: some_error"),
		},
		{
			name: "unknown error code",
			prepareArgs: func() (*DownloadRequest, *DownloadResponse) {
				methodReq := NewDownloadRequest("some_job_id")

				req := v1.NewDownloadRequest(string(methodReq.Id()))
				contentReader := newTestReadCloser("some_content")
				resp := v1.NewResponse(contentReader, v1.JsonContentType)
				s.client.EXPECT().
					SendDownloadRequest(req).
					Return(resp, nil)

				errorResp := v12.NewResponse(88888, "some_error_message")
				s.decoder.EXPECT().
					ToErrorResponse(contentReader).
					Return(errorResp, nil)

				return methodReq, nil
			},
			wantErr: fmt.Errorf("map error code: unknown code: %d", 88888),
		},
	}

	for _, test := range tests {
		test := test
		s.Run(test.name, func() {
			args, wantOut := test.prepareArgs()
			out, err := s.api.Download(args)
			if test.wantErr != nil {
				s.Nil(out)
				s.EqualError(err, test.wantErr.Error())
			} else {
				s.Nil(err)
				s.Equal(wantOut, out)
			}
		})
	}
}

func TestVanceAISuite_Unit(t *testing.T) {
	suite.Run(t, new(vanceAISuite))
}

func (s *vanceAISuite) SetupSuite() {
	s.ctrl = gomock.NewController(s.T())
	s.decoder = mock.NewMockResponseDecoder(s.ctrl)
	s.client = mock.NewMockClient(s.ctrl)
	s.jConfigEncoder = mock.NewMockJConfigEncoder(s.ctrl)

	s.api = &vanceAI{
		client:         s.client,
		respDecoder:    s.decoder,
		jConfigEncoder: s.jConfigEncoder,
	}
}

func (s *vanceAISuite) TearDownSuite() {
	s.ctrl.Finish()
}
