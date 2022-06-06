package v1

import (
	"github.com/golang/mock/gomock"
	"integration-vanceai/internal/filesystem"
	"integration-vanceai/internal/vanceai/v1/image"
	"io"
	"testing"
)

func TestComixifier_Turn_Unit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vanceAI := NewMockVanceAI(ctrl)
	comixifier := NewComixifier(vanceAI, nil) //TODO

	type testCase struct {
		name    string
		prepare func() (filesystem.File, io.ReadCloser)
		wantErr error
	}
	tests := []testCase{
		{
			name: "correct",
			prepare: func() (filesystem.File, io.ReadCloser) {
				img := filesystem.NewMockFile(ctrl)
				wantContent := "some_result"

				uploadResp := NewUploadResponse("some_uid")
				uploadReq := NewUploadRequest("ai", img)
				vanceAI.EXPECT().
					Upload(uploadReq).
					Return(uploadResp, nil)

				transformResp := NewTransformResponse("some_job_id", JobStatusProcess)
				processors := []image.Processor{image.NewCartoonizer()}
				transformReq := NewTransformRequest(uploadResp.Uid(), processors)
				vanceAI.EXPECT().
					Transform(transformReq).
					Return(transformResp, nil)

				progressReq := NewProgressRequest(transformResp.id)
				vanceAI.EXPECT().
					Progress(progressReq).
					Return(NewProgressResponse(JobStatusProcess, 11111), nil).
					Times(7)
				vanceAI.EXPECT().
					Progress(progressReq).
					Return(NewProgressResponse(JobStatusFinish, 11111), nil)

				vanceAI.EXPECT().
					Download(NewDownloadRequest(transformResp.id)).
					Return(NewDownloadResponse(newTestReadCloser(wantContent)), nil)

				return img, newTestReadCloser(wantContent)
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			args, wantOut := test.prepare()

			out, err := comixifier.Turn(args)
			if test.wantErr == nil {
				if err != nil {
					t.Logf("error got: %s; expected: <nil>", err.Error())
					t.FailNow()
				}

				defer wantOut.Close()
				wantContent, _ := io.ReadAll(wantOut)

				if out == nil {
					t.Logf("out is nil; expected: %s", wantContent)
					t.FailNow()
				}

				defer out.Close()
				outContent, _ := io.ReadAll(out)

				if string(wantContent) != string(outContent) {
					t.Logf("out got: %s; expected: %s", outContent, wantContent)
					t.FailNow()
				}
			} else {
				if err == nil {
					t.Logf("error is nil; expected: %s", test.wantErr.Error())
					t.FailNow()
				}
				if err.Error() != test.wantErr.Error() {
					t.Logf("error got: %s; expected: %s", err.Error(), test.wantErr.Error())
					t.FailNow()
				}
			}
		})
	}
}
