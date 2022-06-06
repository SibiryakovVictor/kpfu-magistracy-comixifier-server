package v1

//go:generate mockgen -destination ./v1_mock.go -package v1 --build_flags=--mod=mod integration-vanceai/internal/vanceai/v1 VanceAI

import (
	"comixifier/internal/vanceai/filesystem"
	"comixifier/internal/vanceai/logger"
	"comixifier/internal/vanceai/vanceai/v1/image"
	"fmt"
	"io"
	"time"
)

type Comixifier struct {
	vanceAI VanceAI
	logger  logger.Logger
}

func NewComixifier(vanceAI VanceAI, logger logger.Logger) *Comixifier {
	return &Comixifier{vanceAI: vanceAI, logger: logger}
}

func (c *Comixifier) Turn(img filesystem.File) (io.ReadCloser, error) {
	c.logger.Info("call Upload", nil)
	uploadReq := NewUploadRequest("ai", img)
	uploadResp, err := c.vanceAI.Upload(uploadReq)
	if err != nil {
		c.logger.Error("Upload error", map[string]interface{}{"msg": err.Error()})
		return nil, fmt.Errorf("upload request: %w", err)
	}

	processors := []image.Processor{image.NewCartoonizer()}
	transformReq := NewTransformRequest(uploadResp.Uid(), processors)
	c.logger.Info("call Transform", map[string]interface{}{"uid": transformReq.uid})
	transformResp, err := c.vanceAI.Transform(transformReq)
	if err != nil {
		c.logger.Error("Transform error", map[string]interface{}{"msg": err.Error()})
		return nil, fmt.Errorf("transform request: %w", err)
	}

	progressCount := 0
	progressReq := NewProgressRequest(transformResp.id)
	status := transformResp.status
	for status == JobStatusProcess || status == JobStatusWaiting {
		if progressCount == 15 {
			return nil, fmt.Errorf("progress request limit expired")
		}

		timer := time.NewTimer(10 * time.Second)
		<-timer.C

		c.logger.Info("call Progress", map[string]interface{}{"jobId": progressReq.id})
		progressCount++
		progressResp, err := c.vanceAI.Progress(progressReq)
		if err != nil {
			c.logger.Error("Progress error", map[string]interface{}{"msg": err.Error()})
			return nil, fmt.Errorf("progress request: %w", err)
		}

		status = progressResp.Status()
	}

	c.logger.Info("got job status", map[string]interface{}{"status": status.String()})
	switch status {
	case JobStatusFinish:
		downloadReq := NewDownloadRequest(transformResp.id)
		c.logger.Info("call Download", map[string]interface{}{"jobId": downloadReq.id})
		downloadResp, err := c.vanceAI.Download(downloadReq)
		if err != nil {
			return nil, fmt.Errorf("download request: %w", err)
		}

		return downloadResp.ImgContent(), nil
	case JobStatusFatal:
		return nil, fmt.Errorf("got fatal job status")
	default:
		return nil, fmt.Errorf("unexpected job status: %d, expect 'finish' or 'fatal'", status)
	}
}
