package commands

import (
	"bytes"
	"context"
	"fmt"
	"image/gif"
	"image/png"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/otiai10/amesh-bot/slack"
	"github.com/otiai10/amesh/lib/amesh"
)

// AmeshCommand ...
type AmeshCommand struct{}

// Match ...
func (cmd AmeshCommand) Match(payload *slack.Payload) bool {
	if len(payload.Ext.Words) == 0 {
		return true
	}
	// TODO: これ、他のコマンドが "-a" を持ってたら、こっちがハイジャックしちゃうでしょ
	if payload.Ext.Words.Flag("-a") {
		return true
	}
	return false
}

// Handle ...
func (cmd AmeshCommand) Handle(ctx context.Context, payload *slack.Payload) *slack.Message {

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return wrapError(payload, err)
	}
	now := time.Now().In(tokyo)

	if payload.Ext.Words.Flag("-a") {
		return cmd.ameshAnimated(ctx, payload, now)
	}

	return cmd.ameshNow(ctx, payload, now)
}

func (cmd AmeshCommand) ameshNow(ctx context.Context, payload *slack.Payload, now time.Time) *slack.Message {
	entry := amesh.GetEntry(now)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return wrapError(payload, err)
	}
	defer client.Close()

	bname := os.Getenv("GOOGLE_STORAGE_BUCKET_NAME")
	bucket := client.Bucket(bname)
	datetime := entry.Time.Format("2006-0102-1504")
	fname := fmt.Sprintf("%s.png", datetime)
	furl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bname, fname)
	obj := bucket.Object(fname)

	attrs, err := obj.Attrs(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return wrapError(payload, err)
	}

	// すでにあるのでURLだけ返す
	if attrs != nil && attrs.Size > 0 {
		return &slack.Message{
			Channel: payload.Event.Channel,
			Blocks:  []slack.Block{{Type: "image", ImageURL: furl, AltText: datetime}},
		}
	}

	// 画像の取得と合成
	img, err := entry.GetImage(true, true)
	if err != nil {
		return wrapError(payload, err)
	}
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, img); err != nil {
		return wrapError(payload, err)
	}

	// GoogleStorageへのアップロード
	writer := obj.NewWriter(ctx)
	if _, err = writer.Write(buf.Bytes()); err != nil {
		return wrapError(payload, err)
	}
	if err = writer.Close(); err != nil {
		return wrapError(payload, err)
	}

	return &slack.Message{
		Channel: payload.Event.Channel,
		Blocks:  []slack.Block{{Type: "image", ImageURL: furl, AltText: datetime}},
	}

}

func (cmd AmeshCommand) ameshAnimated(ctx context.Context, payload *slack.Payload, now time.Time) *slack.Message {
	entries := amesh.GetEntries(now.Add(-40*time.Minute), now)

	// {{{ TODO: GCSへのアップロード部分をDRYにするべき
	client, err := storage.NewClient(ctx)
	if err != nil {
		return wrapError(payload, err)
	}
	defer client.Close()

	bname := os.Getenv("GOOGLE_STORAGE_BUCKET_NAME")
	bucket := client.Bucket(bname)
	datetime := entries[0].Time.Format("2006-0102-1504")
	fname := fmt.Sprintf("%s.gif", datetime)
	furl := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bname, fname)
	obj := bucket.Object(fname)

	attrs, err := obj.Attrs(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return wrapError(payload, err)
	}

	// すでにあるのでURLだけ返す
	if attrs != nil && attrs.Size > 0 {
		return &slack.Message{
			Channel: payload.Event.Channel,
			Blocks:  []slack.Block{{Type: "image", ImageURL: furl, AltText: datetime}},
		}
	}
	// }}}

	g, err := entries.ToGif(500, true)
	if err != nil {
		return wrapError(payload, err)
	}
	buf := bytes.NewBuffer(nil)
	if err := gif.EncodeAll(buf, g); err != nil {
		return wrapError(payload, err)
	}

	// GoogleStorageへのアップロード
	writer := obj.NewWriter(ctx)
	if _, err = writer.Write(buf.Bytes()); err != nil {
		return wrapError(payload, err)
	}
	if err = writer.Close(); err != nil {
		return wrapError(payload, err)
	}

	return &slack.Message{
		Channel: payload.Event.Channel,
		Blocks:  []slack.Block{{Type: "image", ImageURL: furl, AltText: datetime}},
	}

}

// Help ...
func (cmd AmeshCommand) Help(payload *slack.Payload) *slack.Message {
	return &slack.Message{
		Channel: payload.Event.Channel,
		Text:    "デフォルトのアメッシュコマンド\n```@amesh [-a]```",
	}
}
