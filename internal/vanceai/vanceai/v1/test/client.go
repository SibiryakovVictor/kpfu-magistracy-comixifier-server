package test

import (
	"comixifier/internal/vanceai/config"
	"comixifier/internal/vanceai/filesystem/local"
	v1http "comixifier/internal/vanceai/http/vanceai/v1"
	v1json "comixifier/internal/vanceai/json/vanceai/v1"
	"encoding/json"
	"fmt"
	"github.com/jaswdr/faker"
	"io"
	"testing"
)

type ClientSuite struct {
	t      *testing.T
	client v1http.Client
}

func NewClientSuite(t *testing.T, client v1http.Client) *ClientSuite {
	return &ClientSuite{
		t:      t,
		client: client,
	}
}

func (s *ClientSuite) TestSendUploadRequest() {
	checkError(s.t, config.Setup(), "setup config")

	img := faker.New().LoremFlickr().Image(10, 10, []string{}, "", false)
	defer img.Close()
	_, err := img.Seek(0, 0)
	if err != nil {
		checkError(s.t, err, "seek file to begin")
	}
	fileImg, err := local.WrapFile(img)
	checkError(s.t, err, "wrap file")

	req := v1http.NewUploadRequest("ai", fileImg)

	resp, err := s.client.SendUploadRequest(req)
	checkError(s.t, err, "send Upload request")

	defer resp.Close()
	bodyJSON, err := io.ReadAll(resp)
	checkError(s.t, err, "read response body")

	body := make(map[string]interface{})
	err = json.Unmarshal(bodyJSON, &body)
	checkError(s.t, err, "unmarshal json body")

	codeRaw, ok := body["code"]
	if !ok {
		checkError(s.t, fmt.Errorf("no 'code' field"), "response body")
	}
	code, ok := codeRaw.(float64)
	if !ok {
		checkError(s.t, fmt.Errorf("'code' field is not int"), "response body")
	}

	if v1json.IsErrorCode(int(code)) {
		checkError(s.t, fmt.Errorf("got error code: %d", int(code)), "response body")
	}

	dataRaw, ok := body["data"]
	if !ok {
		checkError(s.t, fmt.Errorf("no 'data' field"), "response body")
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		checkError(s.t, fmt.Errorf("'data' field is not object"), "response body")
	}

	uidRaw, ok := data["uid"]
	if !ok {
		checkError(s.t, fmt.Errorf("no 'data.uid' field"), "response body")
	}
	uid, ok := uidRaw.(string)
	if !ok {
		checkError(s.t, fmt.Errorf("'data.uid' field is not string"), "response body")
	}

	if uid == "" {
		checkError(s.t, fmt.Errorf("got empty 'data.uid'"), "response body")
	}
}

func checkError(t *testing.T, err error, context string) {
	if err != nil {
		t.Logf("%s: %s\n", context, err.Error())
		t.FailNow()
	}
}
