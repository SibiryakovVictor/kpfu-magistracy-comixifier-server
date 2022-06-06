package cutout

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
)

type Cutout struct {
}

func NewCutout() *Cutout {
	return &Cutout{}
}

func (c *Cutout) Do(imgData io.Reader) (io.Reader, error) {
	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(
			`form-data; name="%s"; filename="%s"`,
			"file", "in.png",
		),
	)
	h.Set("Content-Type", "image/png")
	fileWriter, err := bodyWriter.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("create multipart section for image file: %w", err)
	}
	_, err = io.Copy(fileWriter, imgData)
	if err != nil {
		return nil, fmt.Errorf("copy image file: %w", err)
	}
	err = bodyWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("close multipart body: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://www.cutout.pro/api/v1/cartoonSelfie?cartoonType=5",
		bodyBuf,
	)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	req.Header.Set("APIKEY", os.Getenv("CUTOUT_API_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		defer resp.Body.Close()
		errJsonData, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cutout error: %s", string(errJsonData))
	}

	return resp.Body, nil
}
