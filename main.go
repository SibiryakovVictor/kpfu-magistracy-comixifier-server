package main

import (
	"bytes"
	"comixifier/internal"
	"comixifier/internal/cutout"
	"comixifier/internal/face2comics"
	"comixifier/internal/vanceai"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	redisEndpoint := os.Getenv("STATE_STORAGE_ENDPOINT")
	if redisEndpoint == "" {
		panic("empty env STATE_STORAGE_ENDPOINT")
	}

	stateStorage := redis.NewClient(&redis.Options{
		Addr: redisEndpoint,
	})
	defer stateStorage.Close()

	minioClient, err := getMinio()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		rawReqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("download: read req body: %s\n", err.Error())
			return
		}

		reqBody := map[string]string{
			"transformId": "",
		}
		err = json.Unmarshal(rawReqBody, &reqBody)
		if err != nil {
			log.Printf("download: unmarshal req body from json: %s\n", err.Error())
			return
		}

		imgFilePath, err := stateStorage.Get(context.TODO(), reqBody["transformId"]+"-file").Result()
		if err != nil {
			log.Printf("download: img file path by uid not found: %s\n", err.Error())
			return
		}

		imgFile, err := minioClient.GetObject(
			context.TODO(),
			"test",
			imgFilePath,
			minio.GetObjectOptions{},
		)
		if err != nil {
			log.Printf("download: get file from storage: %s\n", err.Error())
			return
		}
		defer imgFile.Close()

		_, err = io.Copy(w, imgFile)
		if err != nil {
			log.Printf("download: copy file from storage to response: %s\n", err.Error())
			return
		}
		w.Header().Set("Content-Type", "image/png")

		_, err = io.Copy(w, imgFile)
		if err != nil {
			log.Printf("download: copy result image data to response: %s\n", err.Error())
			return
		}
	})

	http.HandleFunc("/progress", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		rawReqBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("progress: read req body: %s\n", err.Error())
			return
		}

		reqBody := map[string]string{
			"transformId": "",
		}
		err = json.Unmarshal(rawReqBody, &reqBody)
		if err != nil {
			log.Printf("progress: unmarshal req body from json: %s\n", err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		status, err := stateStorage.Get(ctx, reqBody["transformId"]+"-status").Result()
		if err != nil {
			log.Printf("progress: get status from state storage: %s\n", err.Error())
			return
		}

		comixifyErr, err := stateStorage.Get(ctx, reqBody["transformId"]+"-error").Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("progress: get transform error from state storage: %s\n", err.Error())
			return
		}

		respBody := map[string]interface{}{
			"status": status,
			"error":  comixifyErr,
		}

		jsonRespBody, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("progress: marshal resp body to json: %s\n", err.Error())
			return
		}

		w.Write(jsonRespBody)
	})

	http.HandleFunc("/transform", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var comixifier internal.Comixifier
		switch r.Header.Get("Comixifier-Name") {
		case "face2comics":
			comixifier = face2comics.NewFace2Comics()
		case "VanceAI":
			comixifier = vanceai.NewVanceAI()
		case "cutout":
			comixifier = cutout.NewCutout()
		default:
			log.Printf("transform: generate uuid: %s\n", err.Error())
			return
		}

		transformId, err := uuid.NewUUID()
		if err != nil {
			log.Printf("transform: generate uuid: %s\n", err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		stateStorage.Set(ctx, transformId.String()+"-status", "WAIT", 10*time.Minute)

		imgBuf := new(bytes.Buffer)
		_, err = io.Copy(imgBuf, r.Body)
		if err != nil {
			log.Printf("transform: copy image from req to buf: %s\n", err.Error())
			return
		}

		go func(comixifier internal.Comixifier, imgData *bytes.Buffer, transformId uuid.UUID) {
			resultImgData, err := comixifier.Do(imgData)
			if err != nil {
				log.Printf("transform: comixify image: %s\n", err.Error())
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				stateStorage.Set(ctx, transformId.String()+"-status", "FATAL", 10*time.Minute)
				stateStorage.Set(ctx, transformId.String()+"-error", err.Error(), 10*time.Minute)
				return
			}

			imgData.Reset()

			imgFile, err := os.CreateTemp("", fmt.Sprintf(
				"img_%s_%d_*.png",
				transformId.String(),
				time.Now().Unix(),
			))
			if err != nil {
				log.Printf("transform: create temp img file: %s\n", err.Error())
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				stateStorage.Set(ctx, transformId.String()+"-status", "FATAL", 10*time.Minute)
				stateStorage.Set(ctx, transformId.String()+"-error", err.Error(), 10*time.Minute)
				return
			}
			defer os.Remove(imgFile.Name())
			defer imgFile.Close()

			_, err = io.Copy(imgFile, resultImgData)
			if err != nil {
				log.Printf("transform: copy result image data to temp img file: %s\n", err.Error())
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				stateStorage.Set(ctx, transformId.String()+"-status", "FATAL", 10*time.Minute)
				stateStorage.Set(ctx, transformId.String()+"-error", err.Error(), 10*time.Minute)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			uploadInfo, err := minioClient.FPutObject(
				ctx,
				"test",
				filepath.Base(imgFile.Name()),
				imgFile.Name(),
				minio.PutObjectOptions{
					ContentType: "image/png",
				},
			)
			if err != nil {
				log.Printf("transform: upload image to image storage: %s\n", err.Error())
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				stateStorage.Set(ctx, transformId.String()+"-status", "FATAL", 10*time.Minute)
				stateStorage.Set(ctx, transformId.String()+"-error", err.Error(), 10*time.Minute)
				return
			}

			stateStorage.Set(ctx, transformId.String()+"-file", uploadInfo.Key, 10*time.Minute)
			stateStorage.Set(ctx, transformId.String()+"-status", "FINISH", 10*time.Minute)
		}(comixifier, imgBuf, transformId)

		respBody := map[string]interface{}{
			"transformId": transformId.String(),
		}

		jsonRespBody, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("transform: marshal resp body to json: %s\n", err.Error())
			return
		}

		w.Write(jsonRespBody)
	})

	err = http.ListenAndServe(":9001", nil)
	if err != nil {
		panic(err)
	}
}

func getMinio() (*minio.Client, error) {
	endpoint := os.Getenv("IMAGE_STORAGE_ENDPOINT")
	if endpoint == "" {
		return nil, fmt.Errorf("empty env IMAGE_STORAGE_ENDPOINT")
	}

	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"

	return minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	})
}

func tryMinio() {
	endpoint := "127.0.0.1:9501"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		//Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectCh := minioClient.ListObjects(ctx, "test", minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(object)
	}

	fmt.Println("success")
}
