package config

import (
	"fmt"
	"github.com/jessevdk/go-flags"
)

var cfg *Config

func Setup() error {
	cfg = &Config{}
	parser := flags.NewParser(cfg, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	return nil
}

func ApiVanceAI() *apiVanceAI {
	return cfg.VanceAI
}

type Config struct {
	TestMode interface{} `short:"t" hidden:"true"`
	VanceAI  *apiVanceAI
}

type apiVanceAI struct {
	ApiToken     string `long:"vanceai-api-token" description:"token for making vanceai api requests" env:"APP_VANCEAI_API_TOKEN" required:"true"`
	UploadURL    string `long:"vanceai-api-upload-url" description:"url to call Upload endpoint" env:"APP_VANCEAI_UPLOAD_URL" required:"true"`
	TransformURL string `long:"vanceai-api-transform-url" description:"url to call Transform endpoint" env:"APP_VANCEAI_TRANSFORM_URL" required:"true"`
	ProgressURL  string `long:"vanceai-api-progress-url" description:"url to call Progress endpoint" env:"APP_VANCEAI_PROGRESS_URL" required:"true"`
	DownloadURL  string `long:"vanceai-api-download-url" description:"url to call Download endpoint" env:"APP_VANCEAI_DOWNLOAD_URL" required:"true"`
}
