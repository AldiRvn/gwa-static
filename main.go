package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"gwa-static/config"
	"gwa-static/util"

	_ "github.com/glebarez/go-sqlite"
	"github.com/logdyhq/logdy-core/logdy"
	"github.com/logrusorgru/aurora/v4"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var client *whatsmeow.Client

func eventHandler(evt interface{}) {
	ctx := context.Background()
	switch msg := evt.(type) {
	case *events.Message:
		msgByte, _ := json.Marshal(msg)
		fmt.Printf("msgByte: \n%v\n", aurora.Green(string(msgByte)))

		conversation, isReplyFile := msg.Message.GetConversation(), false
		if extend := msg.Message.ExtendedTextMessage; conversation == "" && extend != nil && extend.Text != nil {
			conversation = *extend.Text
		}
		if imgMsg := msg.Message.ImageMessage; conversation == "" && imgMsg != nil && imgMsg.Caption != nil {
			conversation = *imgMsg.Caption
		}
		fmt.Println("Received a message!", conversation)

		replyInfo := &waE2E.ContextInfo{StanzaID: &msg.Info.ID, Participant: proto.String(msg.Info.MessageSource.Sender.String()), QuotedMessage: msg.Message}
		reply := &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				ContextInfo: replyInfo,
				Text:        &conversation,
			},
		}
		if msg.Message.ImageMessage != nil {
			fileByte, err := client.Download(ctx, msg.Message.ImageMessage)
			if err != nil {
				log.Println(err)
			} else {
				uploadImage, err := client.Upload(ctx, fileByte, whatsmeow.MediaImage)
				if err != nil {
					log.Println(err)
				} else {
					mimeType := "image/png"
					isReplyFile = true
					reply.ImageMessage = &waE2E.ImageMessage{
						ContextInfo: replyInfo,
						Caption:     &conversation, Mimetype: &mimeType,
						URL: &uploadImage.URL, DirectPath: &uploadImage.DirectPath, MediaKey: uploadImage.MediaKey,
						FileSHA256: uploadImage.FileSHA256, FileEncSHA256: uploadImage.FileEncSHA256, FileLength: &uploadImage.FileLength,
					}
				}
			}
		}

		if isReplyFile {
			//? Jika reply extendedTextMessage dan ImageMessage ga nil maka reply tidak diterima user,
			//? jadi kalau reply selain text maka extendedTextMessagenya buat nil
			reply.ExtendedTextMessage = nil
		}

		replyByte, _ := json.Marshal(reply)
		fmt.Printf("replyByte: \n%v\n", aurora.Yellow(string(replyByte)))

		client.MarkRead(ctx, []types.MessageID{msg.Info.ID}, time.Now(), msg.Info.Sender, msg.Info.Sender)
		client.SendChatPresence(ctx, msg.Info.Sender, types.ChatPresenceComposing, types.ChatPresenceMediaText)
		client.SendMessage(ctx, msg.Info.Chat, reply)
		client.SendChatPresence(ctx, msg.Info.Sender, types.ChatPresencePaused, types.ChatPresenceMediaText)

		config.Logger.Log(logdy.Fields{
			"date":         time.Now().Format(time.DateTime),
			"type":         "wa",
			"msg":          "incoming new message",
			"conversation": conversation,
			"waMsgId":      msg.Info.ID,
			"waSender":     msg.Info.SenderAlt,
			"waPushName":   msg.Info.PushName,
			"rawEventWa":   msg,
		})
	}
}

func main() {
	config.Logger.Log(logdy.Fields{"date": time.Now().Format(time.DateTime), "type": "app", "msg": "started"})

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	ctx := context.Background()
	container, err := sqlstore.New(ctx,
		"sqlite", config.GENERATED_FOLDER+"/wa.db?_pragma=foreign_keys(1)", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}

	// clientLog := waLog.Stdout("Client", "DEBUG", true)
	// client = whatsmeow.NewClient(deviceStore, clientLog) //? default logger

	logger := util.NewLogger("http://0.0.0.0:3100", "gwa-static", slog.LevelDebug)
	client = whatsmeow.NewClient(deviceStore, logger) //? custom logger
	logger.Infof("Running App.")

	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR code:", aurora.Green(evt.Code))
				err := qrcode.WriteFile(evt.Code, qrcode.Medium, 256, config.GENERATED_FOLDER+"/qr.png")
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		client.SendPresence(ctx, types.PresenceAvailable)

		fmt.Printf("client.Store.PushName: %v\n", aurora.Green(client.Store.PushName))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	client.SendPresence(ctx, types.PresenceUnavailable)
	client.Disconnect()
}
