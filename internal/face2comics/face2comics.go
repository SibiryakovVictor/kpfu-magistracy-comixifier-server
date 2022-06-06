package face2comics

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Face2Comics struct {
}

func NewFace2Comics() *Face2Comics {
	return &Face2Comics{}
}

func (f *Face2Comics) Do(imgData io.Reader) (io.Reader, error) {
	imgFile, err := os.Create("in.png")
	if err != nil {
		return nil, fmt.Errorf("create image file in.png: %w", err)
	}
	defer imgFile.Close()
	_, err = io.Copy(imgFile, imgData)
	if err != nil {
		return nil, fmt.Errorf("copy image data to file: %w", err)
	}

	appId := os.Getenv("FACE2COMICS_APP_ID")
	if appId == "" {
		return nil, fmt.Errorf("empty env FACE2COMICS_APP_ID")
	}
	err = os.Setenv("APP_ID", appId)
	if err != nil {
		return nil, fmt.Errorf("set env APP_ID: %w", err)
	}

	appHash := os.Getenv("FACE2COMICS_APP_HASH")
	if appHash == "" {
		return nil, fmt.Errorf("empty env FACE2COMICS_APP_HASH")
	}
	err = os.Setenv("APP_HASH", appHash)
	if err != nil {
		return nil, fmt.Errorf("set env APP_HASH: %w", err)
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	defer func() { _ = log.Sync() }()
	// No graceful shutdown.
	ctx := context.Background()

	phone := os.Getenv("FACE2COMICS_PHONE")
	if phone == "" {
		return nil, fmt.Errorf("empty env FACE2COMICS_PHONE")
	}
	// Setting up authentication flow helper based on terminal auth.
	flow := auth.NewFlow(
		termAuth{phone: phone},
		auth.SendCodeOptions{},
	)

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		Logger: log,
	})
	if err != nil {
		return nil, fmt.Errorf("create telegram client: %w", err)
	}

	var resultImgData io.Reader
	err = client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		log.Info("AUTH: SUCCESS")

		// * Resolve face2comicsbot to get user id and access hash
		//err = resolveBot(ctx, client, log)
		//if err != nil {
		//	return err
		//}

		// * Upload and send image
		msgId, err := sendImage(ctx, client, log, "in.png")
		if err != nil {
			return fmt.Errorf("send image: %w", err)
		}

		log.Info("SEND IMAGE: SUCCESS")

		var resultImgBytes []byte
		errChan := make(chan error)
		go func() {
			timer := time.NewTimer(30 * time.Second)
			<-timer.C

			messagesGeneral, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer: &tg.InputPeerUser{
					UserID:     1319170847,
					AccessHash: 7054378819286729772,
				},
				MinID: msgId,
				MaxID: msgId + 9,
			})
			if err != nil {
				errChan <- fmt.Errorf("get messages history: %w", err)
				return
			}

			log.Info("MESSAGES GET HISTORY: SUCCESS")

			switch messagesGeneral.(type) {
			case *tg.MessagesMessagesSlice:
				log.Info("TYPE - *tg.MessagesMessagesSlice")
				messages := messagesGeneral.(*tg.MessagesMessagesSlice)
				for i, messageGeneral := range messages.Messages {
					switch messageGeneral.(type) {
					case *tg.Message:
						message := messageGeneral.(*tg.Message)
						switch message.Media.(type) {
						case *tg.MessageMediaPhoto:
							mediaPhoto := message.Media.(*tg.MessageMediaPhoto)
							switch mediaPhoto.Photo.(type) {
							case *tg.Photo:
								photo := mediaPhoto.Photo.(*tg.Photo)

								uploadGetFileResp, err := client.API().UploadGetFile(ctx, &tg.UploadGetFileRequest{
									Location: &tg.InputPhotoFileLocation{
										ID:            photo.ID,
										AccessHash:    photo.AccessHash,
										FileReference: photo.FileReference,
										ThumbSize:     photo.Sizes[0].GetType(),
									},
									Limit: 512 * 1024,
								})
								if err != nil {
									errChan <- fmt.Errorf("uplaod get file: %w", err)
									return
								}

								switch uploadGetFileResp.(type) {
								case *tg.UploadFile:
									uploadFile := uploadGetFileResp.(*tg.UploadFile)
									resultImgBytes = uploadFile.Bytes
									errChan <- nil
									return
								}

								log.Info(fmt.Sprintf("GOT UPLOAD RESPONSE %d", i),
									zap.Any("upload file", uploadGetFileResp))
							}
						}
					}

					log.Info(fmt.Sprintf("GOT MESSAGE %d", i), zap.Any("message", messageGeneral))
				}
			default:
				errChan <- fmt.Errorf("unknown message type")
				return
			}

			errChan <- nil
		}()

		err = <-errChan
		if err != nil {
			return fmt.Errorf("check message history with face2comics: %w", err)
		}

		resultImgData = bytes.NewReader(resultImgBytes)
		return nil
	})

	return resultImgData, err
}

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("not implemented")
}

func (c noSignUp) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

// termAuth implements authentication via terminal.
type termAuth struct {
	noSignUp
	phone string
}

func (a termAuth) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (a termAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	bytePwd, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePwd)), nil
}

func (a termAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func sendImage(
	ctx context.Context,
	client *telegram.Client,
	log *zap.Logger,
	imgFilePath string,
) (int, error) {
	img, err := os.Open(imgFilePath)
	if err != nil {
		return 0, fmt.Errorf("open in.png: %w", err)
	}

	nBytes, nChunks := int64(0), 0
	r := bufio.NewReader(img)
	buf := make([]byte, 0, 16*1024)
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return 0, err
		}

		ok, err := client.API().UploadSaveFilePart(ctx, &tg.UploadSaveFilePartRequest{
			FileID:   777888999000,
			FilePart: nChunks,
			Bytes:    buf,
		})
		if err != nil {
			return 0, fmt.Errorf("upload file: %w", err)
		}
		if !ok {
			return 0, fmt.Errorf("can't upload file")
		}

		nChunks++
		nBytes += int64(len(buf))

		// process buf
		if err != nil && err != io.EOF {
			return 0, err
		}

		fmt.Printf("chunks: %d, bytes: %d\n", nChunks, nBytes)
	}

	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 99999999999

	upd, err := client.API().MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer: &tg.InputPeerUser{
			UserID:     1319170847,
			AccessHash: 7054378819286729772,
		},
		ReplyToMsgID: 0,
		Media: &tg.InputMediaUploadedPhoto{
			File: &tg.InputFile{
				ID:    777888999000,
				Parts: nChunks,
				Name:  "in.png",
			},
		},
		Message:  "my message!",
		RandomID: int64(rand.Intn(max-min+1) + min),
	})

	if err != nil {
		return 0, fmt.Errorf("send message: %w", err)
	}

	log.Info("SEND MEDIA RESULT", zap.Any("updates", upd))

	switch upd.(type) {
	case *tg.Updates:
		updates := upd.(*tg.Updates)
		for _, update := range updates.Updates {
			switch update.(type) {
			case *tg.UpdateMessageID:
				updateMsg := update.(*tg.UpdateMessageID)
				return updateMsg.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("unknown updates type")
}

func resolveBot(ctx context.Context, client *telegram.Client, log *zap.Logger) error {
	resolv, err := client.API().ContactsResolveUsername(ctx, "face2comicsbot")
	if err != nil {
		return fmt.Errorf("resolve contract @face2comicsbot: %w", err)
	}

	log.Info("RESOLVE SUCCESS", zap.Any("resolve", resolv))

	return nil
}
