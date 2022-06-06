package builtin

import (
	"bytes"
	"comixifier/internal/vanceai/config"
	v1 "comixifier/internal/vanceai/http/vanceai/v1"
	clienttest "comixifier/internal/vanceai/vanceai/v1/test"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SendUploadRequest_Func(t *testing.T) {
	err := config.Setup()
	if err != nil {
		t.Logf("can't setup config: %s", err.Error())
		t.FailNow()
	}

	endpoints := NewEndpoints(config.ApiVanceAI().UploadURL, "", "")

	clientSuite := clienttest.NewClientSuite(t, NewClient(config.ApiVanceAI().ApiToken, endpoints))
	clientSuite.TestSendUploadRequest()
}

type testReadCloser struct {
	bytes.Buffer
}

func (rc *testReadCloser) Close() error {
	return nil
}

type testHandlerTransform struct {
	testCase         string
	expectedApiToken string
	expectedUid      string
	expectedJConfig  string
}

func (h *testHandlerTransform) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch h.testCase {
	case "correct":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("method is not POST"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		if r.Form.Get("api_token") != h.expectedApiToken {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty api_token"))
			return
		}
		if r.Form.Get("uid") != h.expectedUid {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty uid"))
			return
		}
		if r.Form.Get("jconfig") != h.expectedJConfig {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty jconfig"))
			return
		}

		w.Write([]byte("response from server"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("testCase is not set"))
		if err != nil {
			panic(err)
		}
	}
}

func TestClient_SendTransformRequest_Unit(t *testing.T) {
	th := &testHandlerTransform{}
	ts := httptest.NewServer(th)
	defer ts.Close()

	type testCase struct {
		name      string
		prepare   func()
		argsBuild func() string
		argsCall  func() *v1.TransformRequest
		wantOut   func() io.ReadCloser
		wantErr   error
	}
	tests := []testCase{
		{
			name: "correct",
			prepare: func() {
				th.testCase = "correct"
				th.expectedApiToken = "expected_api_token"
				th.expectedUid = "some_uid"
				th.expectedJConfig = "some_jConfig"
			},
			argsBuild: func() string {
				return "expected_api_token"
			},
			argsCall: func() *v1.TransformRequest {
				jConfig := new(bytes.Buffer)
				jConfig.WriteString("some_jConfig")
				return v1.NewTransformRequest("some_uid", jConfig)
			},
			wantOut: func() io.ReadCloser {
				readCloser := new(testReadCloser)
				readCloser.WriteString("response from server")
				return readCloser
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.prepare()

			wantOutReader := test.wantOut()
			defer wantOutReader.Close()
			wantOut, _ := io.ReadAll(wantOutReader)

			client := NewClient(test.argsBuild(), NewEndpoints("", ts.URL, ""))
			outReader, err := client.SendTransformRequest(test.argsCall())
			if test.wantErr != nil {
				if outReader != nil {
					t.Logf("SendTransformRequest() outReader got: %#v, expected: %#v", outReader, nil)
					t.FailNow()
				}
				if err == nil || err.Error() != test.wantErr.Error() {
					t.Logf("SendTransformRequest() err got: %v, expected: %v", err, test.wantErr)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Logf("SendTransformRequest() err got: %v, expected %v", err, nil)
					t.FailNow()
				}

				if outReader == nil {
					t.Logf("SendTransformRequest() outReader got: %#v, expected: %s", outReader, wantOut)
					t.FailNow()
				}
				defer outReader.Close()
				out, err := io.ReadAll(outReader)
				if err != nil {
					t.Logf("SendTransformRequest() readAll from outReader: %s", err.Error())
					t.FailNow()
				}
				if string(out) != string(wantOut) {
					t.Logf("SendTransformRequest() out got: %s, expected: %s", out, wantOut)
					t.FailNow()
				}
			}
		})
	}
}

type testHandlerProgress struct {
	testCase         string
	expectedApiToken string
	expectedJobId    string
}

func (h *testHandlerProgress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch h.testCase {
	case "correct":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("method is not POST"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		if r.Form.Get("api_token") != h.expectedApiToken {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty api_token"))
			return
		}
		if r.Form.Get("trans_id") != h.expectedJobId {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty trans_id"))
			return
		}

		w.Write([]byte("response from server"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("testCase is not set"))
		if err != nil {
			panic(err)
		}
	}
}

func TestClient_SendProgressRequest_Unit(t *testing.T) {
	th := &testHandlerProgress{}
	ts := httptest.NewServer(th)
	defer ts.Close()

	type testCase struct {
		name             string
		prepareArgsBuild func() (string, *Endpoints)
		prepareArgsCall  func() *v1.ProgressRequest
		wantOut          func() io.ReadCloser
		wantErr          error
	}
	tests := []testCase{
		{
			name: "correct",
			prepareArgsBuild: func() (string, *Endpoints) {
				apiToken := "expected_api_token"

				th.expectedApiToken = apiToken

				return apiToken, NewEndpoints("", "", ts.URL)
			},
			prepareArgsCall: func() *v1.ProgressRequest {
				jobId := "some_job_id"

				th.testCase = "correct"
				th.expectedJobId = jobId

				return v1.NewProgressRequest(jobId)
			},
			wantOut: func() io.ReadCloser {
				readCloser := new(testReadCloser)
				readCloser.WriteString("response from server")
				return readCloser
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wantOutReader := test.wantOut()
			defer wantOutReader.Close()
			wantOut, _ := io.ReadAll(wantOutReader)

			apiToken, endpoints := test.prepareArgsBuild()
			client := NewClient(apiToken, endpoints)

			r := test.prepareArgsCall()
			outReader, err := client.SendProgressRequest(r)
			if test.wantErr != nil {
				if outReader != nil {
					t.Logf("SendProgressRequest() outReader got: %#v, expected: %#v", outReader, nil)
					t.FailNow()
				}
				if err == nil || err.Error() != test.wantErr.Error() {
					t.Logf("SendProgressRequest() err got: %v, expected: %v", err, test.wantErr)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Logf("SendProgressRequest() err got: %v, expected %v", err, nil)
					t.FailNow()
				}

				if outReader == nil {
					t.Logf("SendProgressRequest() outReader got: %#v, expected: %s", outReader, wantOut)
					t.FailNow()
				}
				defer outReader.Close()
				out, err := io.ReadAll(outReader)
				if err != nil {
					t.Logf("SendProgressRequest() readAll from outReader: %s", err.Error())
					t.FailNow()
				}
				if string(out) != string(wantOut) {
					t.Logf("SendProgressRequest() out got: %s, expected: %s", out, wantOut)
					t.FailNow()
				}
			}
		})
	}
}

type testHandlerDownload struct {
	testCase                       string
	expectedApiToken               string
	expectedJobId                  string
	expectedRespContentType        string
	expectedRespContentDisposition string
}

func (h *testHandlerDownload) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch h.testCase {
	case "correct":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("method is not POST"))
			return
		}

		expectedContentType := "application/x-www-form-urlencoded"
		gotContentType := r.Header.Get("Content-Type")
		if gotContentType != expectedContentType {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Content-Type got: %s, expected: %s", gotContentType, expectedContentType)))
			return
		}

		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		if r.Form.Get("api_token") != h.expectedApiToken {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty api_token"))
			return
		}
		if r.Form.Get("trans_id") != h.expectedJobId {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("empty trans_id"))
			return
		}

		w.Header().Set("Content-Disposition", h.expectedRespContentDisposition)
		w.Header().Set("Content-Type", h.expectedRespContentType)
		w.Write([]byte("response from server"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("testCase is not set"))
		if err != nil {
			panic(err)
		}
	}
}

func TestClient_SendDownloadRequest_Unit(t *testing.T) {
	th := &testHandlerDownload{}
	ts := httptest.NewServer(th)
	defer ts.Close()

	type testCase struct {
		name             string
		prepareArgsBuild func() (string, *Endpoints)
		prepareArgsCall  func() *v1.DownloadRequest
		wantOut          func() *v1.Response
		wantErr          func() error
	}
	tests := []testCase{
		{
			name: "correct img content type",
			prepareArgsBuild: func() (string, *Endpoints) {
				apiToken := "expected_api_token"

				th.expectedApiToken = apiToken

				endpoints := NewEndpoints("", "", "")
				endpoints.SetDownload(ts.URL)

				return apiToken, endpoints
			},
			prepareArgsCall: func() *v1.DownloadRequest {
				jobId := "some_job_id"

				th.testCase = "correct"
				th.expectedJobId = jobId

				return v1.NewDownloadRequest(jobId)
			},
			wantOut: func() *v1.Response {
				contentReader := new(testReadCloser)
				contentReader.WriteString("response from server")
				contentType := v1.ImgContentType

				th.expectedRespContentDisposition = `attachment; filename="cartoonize_q5eNr_in.png"`

				return v1.NewResponse(contentReader, contentType)
			},
			wantErr: func() error {
				return nil
			},
		},
		{
			name: "correct json content type",
			prepareArgsBuild: func() (string, *Endpoints) {
				apiToken := "expected_api_token"

				th.expectedApiToken = apiToken

				endpoints := NewEndpoints("", "", "")
				endpoints.SetDownload(ts.URL)

				return apiToken, endpoints
			},
			prepareArgsCall: func() *v1.DownloadRequest {
				jobId := "some_job_id"

				th.testCase = "correct"
				th.expectedJobId = jobId

				return v1.NewDownloadRequest(jobId)
			},
			wantOut: func() *v1.Response {
				contentReader := new(testReadCloser)
				contentReader.WriteString("response from server")
				contentType := v1.JsonContentType

				th.expectedRespContentType = "application/json"
				th.expectedRespContentDisposition = ""

				return v1.NewResponse(contentReader, contentType)
			},
			wantErr: func() error {
				return nil
			},
		},
		{
			name: "unexpected content type",
			prepareArgsBuild: func() (string, *Endpoints) {
				apiToken := "expected_api_token"

				th.expectedApiToken = apiToken

				endpoints := NewEndpoints("", "", "")
				endpoints.SetDownload(ts.URL)

				return apiToken, endpoints
			},
			prepareArgsCall: func() *v1.DownloadRequest {
				jobId := "some_job_id"

				th.testCase = "correct"
				th.expectedJobId = jobId

				return v1.NewDownloadRequest(jobId)
			},
			wantOut: func() *v1.Response {
				contentReader := new(testReadCloser)
				contentReader.WriteString("response from server")
				contentType := v1.UnknownContentType

				th.expectedRespContentType = "random_content_type"
				th.expectedRespContentDisposition = "random_content_disposition"

				return v1.NewResponse(contentReader, contentType)
			},
			wantErr: func() error {
				return nil
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			apiToken, endpoints := test.prepareArgsBuild()
			client := NewClient(apiToken, endpoints)

			wantOut, wantErr := test.wantOut(), test.wantErr()
			r := test.prepareArgsCall()
			resp, err := client.SendDownloadRequest(r)
			if wantErr != nil {
				if resp != nil {
					t.Logf("SendDownloadRequest() resp got: %#v, expected: %#v", resp, nil)
					t.FailNow()
				}
				if err == nil || err.Error() != wantErr.Error() {
					t.Logf("SendDownloadRequest() err got: %v, expected: %v", err, wantErr)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Logf("SendDownloadRequest() err got: %v, expected %v", err, nil)
					t.FailNow()
				}

				if resp == nil {
					t.Logf("SendDownloadRequest() resp got: %#v, expected: %#v", resp, wantOut)
					t.FailNow()
				}

				if wantOut.ContentType() != resp.ContentType() {
					t.Logf(
						"SendDownloadRequest() out.contentType got: %d, expected: %d",
						resp.ContentType(),
						wantOut.ContentType(),
					)
					t.FailNow()
				}

				defer wantOut.Content().Close()
				wantContent, _ := io.ReadAll(wantOut.Content())

				var outContent []byte
				if resp.Content() != nil {
					defer resp.Content().Close()
					outContent, _ = io.ReadAll(resp.Content())
				}

				if string(outContent) != string(wantContent) {
					t.Logf("SendDownloadRequest() out.content got: %s, expected: %s", outContent, wantContent)
					t.FailNow()
				}
			}
		})
	}
}
