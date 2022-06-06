package vanceai

import (
	"comixifier/internal/vanceai/filesystem/local"
	"comixifier/internal/vanceai/http/vanceai/v1/builtin"
	builtin2 "comixifier/internal/vanceai/json/vanceai/v1/builtin"
	zap2 "comixifier/internal/vanceai/logger/zap"
	v1 "comixifier/internal/vanceai/vanceai/v1"
	"fmt"
	"go.uber.org/zap"
	"io"
	"os"
)

type VanceAI struct {
}

func NewVanceAI() *VanceAI {
	return &VanceAI{}
}

func (v *VanceAI) Do(imgData io.Reader) (io.Reader, error) {
	pkgLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	defer pkgLogger.Sync()
	sugar := pkgLogger.Sugar()
	logger := zap2.NewLogger(sugar)

	endpoints := builtin.NewEndpoints(
		"https://api-service.vanceai.com/web_api/v1/upload",
		"https://api-service.vanceai.com/web_api/v1/transform",
		"https://api-service.vanceai.com/web_api/v1/progress",
	)
	endpoints.SetDownload("https://api-service.vanceai.com/web_api/v1/download")
	client := builtin.NewClient(os.Getenv("VANCEAI_API_TOKEN"), endpoints)
	respDecoder := builtin2.NewResponseDecoder()
	jConfigEncoder := builtin2.NewJConfigEncoder()
	vanceAI := v1.NewVanceAI(client, respDecoder, jConfigEncoder)

	comixifier := v1.NewComixifier(vanceAI, logger)

	imgFile, err := os.Create("in.png")
	if err != nil {
		return nil, fmt.Errorf("create image file in.png: %w", err)
	}
	_, err = io.Copy(imgFile, imgData)
	if err != nil {
		imgFile.Close()
		return nil, fmt.Errorf("copy image data to file: %w", err)
	}
	imgFile.Close()

	imgFile, err = os.Open("in.png")
	if err != nil {
		return nil, fmt.Errorf("open image file in.png: %w", err)
	}

	imgWrapFile, err := local.WrapFile(imgFile)
	if err != nil {
		return nil, fmt.Errorf("wrap local file: %w", err)
	}

	return comixifier.Turn(imgWrapFile)
}
