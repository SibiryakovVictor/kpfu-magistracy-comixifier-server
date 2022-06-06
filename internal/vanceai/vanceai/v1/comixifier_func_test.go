package v1

import (
	"comixifier/internal/vanceai/config"
	"comixifier/internal/vanceai/filesystem/local"
	"comixifier/internal/vanceai/http/vanceai/v1/builtin"
	builtin2 "comixifier/internal/vanceai/json/vanceai/v1/builtin"
	zap2 "comixifier/internal/vanceai/logger/zap"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"testing"
)

func TestComixifier_Turn_Func(t *testing.T) {
	checkError(t, config.Setup(), "setup config")

	pkgLogger, err := zap.NewDevelopment()
	checkError(t, err, "create zap logger")
	defer pkgLogger.Sync()
	sugar := pkgLogger.Sugar()
	logger := zap2.NewLogger(sugar)

	endpoints := builtin.NewEndpoints(
		config.ApiVanceAI().UploadURL,
		config.ApiVanceAI().TransformURL,
		config.ApiVanceAI().ProgressURL,
	)
	endpoints.SetDownload(config.ApiVanceAI().DownloadURL)
	client := builtin.NewClient(config.ApiVanceAI().ApiToken, endpoints)
	respDecoder := builtin2.NewResponseDecoder()
	jConfigEncoder := builtin2.NewJConfigEncoder()
	vanceAI := NewVanceAI(client, respDecoder, jConfigEncoder)

	imgFile, err := os.Open("../../../samples/in.png")
	checkError(t, err, "open image file")
	defer imgFile.Close()
	imgLocalFile, _ := local.WrapFile(imgFile)

	comixifier := NewComixifier(vanceAI, logger)
	imgData, err := comixifier.Turn(imgLocalFile)
	checkError(t, err, "comixifier turns image")

	imgOutFile, err := os.Create("../../../samples/out.txt")
	checkError(t, err, "create out image file")
	defer imgOutFile.Close()

	_, err = io.Copy(imgOutFile, imgData)
	checkError(t, err, "copy content to out file")

	log.Println("success")
}

func checkError(t *testing.T, err error, context string) {
	if err != nil {
		t.Logf("%s: %s\n", context, err.Error())
		t.FailNow()
	}
}
