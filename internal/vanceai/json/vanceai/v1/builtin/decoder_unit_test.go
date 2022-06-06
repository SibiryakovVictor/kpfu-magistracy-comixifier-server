package builtin

import (
	"bytes"
	"comixifier/internal/vanceai/json/vanceai/v1"
	"comixifier/internal/vanceai/json/vanceai/v1/builtin/schema"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestResponseDecoder_ToTransformResponse_Unit(t *testing.T) {
	type testCase struct {
		name    string
		args    interface{}
		wantOut *v1.TransformResponse
		wantErr error
	}

	tests := []testCase{
		{
			name: "error response",
			args: &schema.ErrorResponse{
				Code: 10001,
				Msg:  "some error",
			},
			wantOut: v1.NewTransformErrorResponse(10001, "some error"),
			wantErr: nil,
		},
		{
			name: "common response",
			args: schema.TransformResponse{
				Response: schema.Response{
					Code: 200,
				},
				Data: schema.TransformResponseData{
					TransId: "some_trans_job_id",
					Status:  "some_status",
				},
			},
			wantOut: v1.NewTransformResponse(200, "some_trans_job_id", "some_status"),
			wantErr: nil,
		},
	}

	decoder := NewResponseDecoder()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(test.args)

			out, err := decoder.ToTransformResponse(buf)
			if test.wantErr != nil {
				if out != nil {
					t.Logf("ToTransformResponse() out got: %#v, expected %#v", out, nil)
					t.FailNow()
				}
				if err == nil || err.Error() != test.wantErr.Error() {
					t.Logf("ToTransformResponse() err got: %v, expected %v", err, test.wantErr)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Logf("ToTransformResponse() err got: %v, expected %v", err, nil)
					t.FailNow()
				}
				if !reflect.DeepEqual(out, test.wantOut) {
					t.Logf("ToTransformResponse() out got: %#v, expected %#v", out, test.wantOut)
					t.FailNow()
				}
			}
		})
	}
}

func TestResponseDecoder_ToTransformResponse_Errors_Unit(t *testing.T) {
	decoder := NewResponseDecoder()
	t.Run("incorrect json", func(t *testing.T) {
		wantErr := fmt.Errorf("decode to ErrorResponse: invalid character '}' after top-level value")
		out, err := decoder.ToTransformResponse(bytes.NewBufferString("{}}"))
		if out != nil {
			t.Logf("ToTransformResponse() out got: %#v, expected %#v", out, nil)
			t.FailNow()
		}
		if err == nil || err.Error() != wantErr.Error() {
			t.Logf("ToTransformResponse() err got: %v, expected %v", err, wantErr)
			t.FailNow()
		}
	})
}

func TestResponseDecoder_ToProgressResponse_Unit(t *testing.T) {
	type testCase struct {
		name    string
		args    interface{}
		wantOut *v1.ProgressResponse
		wantErr error
	}

	tests := []testCase{
		{
			name: "error response",
			args: &schema.ErrorResponse{
				Code: 10001,
				Msg:  "some error",
			},
			wantOut: v1.NewProgressErrorResponse(10001, "some error"),
			wantErr: nil,
		},
		{
			name: "common response",
			args: schema.ProgressResponse{
				Response: schema.Response{
					Code: 200,
				},
				Data: schema.ProgressResponseData{
					Status:   "some_status",
					Filesize: uint32(12345678),
				},
			},
			wantOut: v1.NewProgressResponse(200, "some_status", uint32(12345678)),
			wantErr: nil,
		},
	}

	decoder := NewResponseDecoder()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(test.args)

			out, err := decoder.ToProgressResponse(buf)
			if test.wantErr != nil {
				if out != nil {
					t.Logf("ToProgressResponse() out got: %#v, expected %#v", out, nil)
					t.FailNow()
				}
				if err == nil || err.Error() != test.wantErr.Error() {
					t.Logf("ToProgressResponse() err got: %v, expected %v", err, test.wantErr)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Logf("ToProgressResponse() err got: %v, expected %v", err, nil)
					t.FailNow()
				}
				if !reflect.DeepEqual(out, test.wantOut) {
					t.Logf("ToProgressResponse() out got: %#v, expected %#v", out, test.wantOut)
					t.FailNow()
				}
			}
		})
	}
}

func TestResponseDecoder_ToErrorResponse_Unit(t *testing.T) {
	type testCase struct {
		name    string
		args    func() io.Reader
		wantOut *v1.Response
		wantErr error
	}

	tests := []testCase{
		{
			name: "correct",
			args: func() io.Reader {
				errorResp := &schema.ErrorResponse{
					Code: 10001,
					Msg:  "some error",
				}
				buf := new(bytes.Buffer)
				json.NewEncoder(buf).Encode(errorResp)

				return buf
			},
			wantOut: v1.NewResponse(10001, "some error"),
			wantErr: nil,
		},
		{
			name: "incorrect json",
			args: func() io.Reader {
				buf := new(bytes.Buffer)
				buf.WriteString("{{}")

				return buf
			},
			wantOut: nil,
			wantErr: fmt.Errorf("decode data to ErrorResponse: invalid character '{' looking for beginning of object key string"),
		},
		{
			name: "incorrect type of error msg",
			args: func() io.Reader {
				errorResp := &schema.ErrorResponse{
					Code: 10001,
					Msg:  111,
				}
				buf := new(bytes.Buffer)
				json.NewEncoder(buf).Encode(errorResp)

				return buf
			},
			wantOut: nil,
			wantErr: fmt.Errorf("get error msg: unknown type of msg: float64"),
		},
	}

	decoder := NewResponseDecoder()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			data := test.args()
			out, err := decoder.ToErrorResponse(data)
			if test.wantErr != nil {
				if out != nil {
					t.Logf("ToErrorResponse() out got: %#v, expected: %#v", out, nil)
					t.FailNow()
				}
				if err == nil || err.Error() != test.wantErr.Error() {
					t.Logf("ToErrorResponse() err got: %v, expected: %v", err, test.wantErr)
					t.FailNow()
				}
			} else {
				if err != nil {
					t.Logf("ToErrorResponse() err got: %v, expected: %v", err, nil)
					t.FailNow()
				}
				if !reflect.DeepEqual(out, test.wantOut) {
					t.Logf("ToErrorResponse() out got: %#v, expected: %#v", out, test.wantOut)
					t.FailNow()
				}
			}
		})
	}
}
