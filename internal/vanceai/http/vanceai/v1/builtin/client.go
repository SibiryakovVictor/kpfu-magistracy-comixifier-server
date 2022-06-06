package builtin

import (
	"bytes"
	v1 "comixifier/internal/vanceai/http/vanceai/v1"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
)

type Client struct {
	apiToken  string
	endpoints *Endpoints
}

func NewClient(apiToken string, endpoints *Endpoints) *Client {
	return &Client{
		apiToken:  apiToken,
		endpoints: endpoints,
	}
}

func (c *Client) SendUploadRequest(methodReq *v1.UploadRequest) (io.ReadCloser, error) {
	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)

	err := bodyWriter.WriteField("api_token", c.apiToken)
	if err != nil {
		return nil, fmt.Errorf("add api_token to request body: %w", err)
	}
	err = bodyWriter.WriteField("job", methodReq.Job())
	if err != nil {
		return nil, fmt.Errorf("add job to request body: %w", err)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(
			`form-data; name="%s"; filename="%s"`,
			"file", methodReq.Image().Name(),
		),
	)
	h.Set("Content-Type", "image/"+filepath.Ext(methodReq.Image().Name())[1:])
	fileWriter, err := bodyWriter.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("create multipart section for image file: %w", err)
	}
	_, err = io.Copy(fileWriter, methodReq.Image().Content())
	if err != nil {
		return nil, fmt.Errorf("copy image file: %w", err)
	}
	err = bodyWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("close multipart body: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.endpoints.upload, bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	client := &http.Client{} //TODO timeouts
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send http request: %w", err)
	}
	return httpResp.Body, nil
}

func (c *Client) SendTransformRequest(methodReq *v1.TransformRequest) (io.ReadCloser, error) {
	jConfigReader := methodReq.JConfig()
	jConfig, err := io.ReadAll(jConfigReader)
	if err != nil {
		return nil, fmt.Errorf("read jConfig: %w", err)
	}

	form := &url.Values{}
	form.Add("api_token", c.apiToken)
	form.Add("uid", methodReq.Uid())
	form.Add("jconfig", string(jConfig))

	httpReq, err := http.NewRequest(http.MethodPost, c.endpoints.transform, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp.Body, nil
}

func (c *Client) SendProgressRequest(methodReq *v1.ProgressRequest) (io.ReadCloser, error) {
	form := &url.Values{}
	form.Add("api_token", c.apiToken)
	form.Add("trans_id", methodReq.JobId())

	httpReq, err := http.NewRequest(http.MethodPost, c.endpoints.progress, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp.Body, nil
}

func (c *Client) SendDownloadRequest(methodReq *v1.DownloadRequest) (*v1.Response, error) {
	form := &url.Values{}
	form.Add("api_token", c.apiToken)
	form.Add("trans_id", methodReq.JobId())

	httpReq, err := http.NewRequest(http.MethodPost, c.endpoints.download, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	methodRespContentType := v1.UnknownContentType
	respContentType := resp.Header.Get("Content-Type")
	respContentDisposition := resp.Header.Get("Content-Disposition")
	if strings.HasPrefix(respContentDisposition, "attachment") {
		methodRespContentType = v1.ImgContentType
	} else if strings.HasPrefix(respContentType, "application/json") {
		methodRespContentType = v1.JsonContentType
	}

	return v1.NewResponse(resp.Body, methodRespContentType), nil
}

type Endpoints struct {
	upload    string
	transform string
	progress  string
	download  string
}

func (e *Endpoints) SetDownload(download string) {
	e.download = download
}

func NewEndpoints(upload string, transform string, progress string) *Endpoints {
	return &Endpoints{
		upload:    upload,
		transform: transform,
		progress:  progress,
	}
}
