package commands

import (
	"bytes"
	"context"
	"fmt"
	"image/gif"
	"image/png"
	"io"
	"os"
	"strings"
	"time"

	"github.com/otiai10/amesh-bot/service"
	"github.com/otiai10/amesh/lib/amesh"
	"github.com/otiai10/largo"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type (
	CloudStorage interface {
		Exists(ctx context.Context, bucket string, name string) (exists bool, err error)
		Upload(ctx context.Context, bucket string, name string, contents []byte) error
		URL(bucket string, name string) string
	}
)

// AmeshCommand ...
type AmeshCommand struct {
	Storage CloudStorage
}

func (cmd AmeshCommand) newFlagSet(animated bool, help io.Writer) *largo.FlagSet {
	fset := largo.NewFlagSet("", largo.ContinueOnError)
	fset.Output = help
	fset.BoolVar(&animated, "animated", false, "GIF画像でタイムラプス表示").Alias("a")
	return fset
}

// Match ...
func (cmd AmeshCommand) Match(event slackevents.AppMentionEvent) bool {
	fset := cmd.newFlagSet(false, Discard)
	fset.Parse(largo.Tokenize(event.Text)[1:])
	return len(fset.Rest()) == 0
}

func (cmd AmeshCommand) Execute(ctx context.Context, client *service.SlackClient, event slackevents.AppMentionEvent) (err error) {

	var animated bool
	help := bytes.NewBuffer(nil)
	fset := cmd.newFlagSet(animated, help)

	tokens := strings.Fields(event.Text)[1:]
	if err := fset.Parse(tokens); err != nil {
		return fmt.Errorf("failed to parse arguments: %v", err)
	}

	msg := service.SlackMsg{Channel: event.Channel}

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("failed to load location: %v", err)
	}
	now := time.Now().In(tokyo)

	switch {
	case fset.HelpRequested():
		msg.Text = fmt.Sprintf("デフォルトのアメッシュコマンド\n```@amesh [-a] [-h]\n%v```", help.String())
	case animated:
		block, err := cmd.ameshAnimated(ctx, now)
		if err != nil {
			return err
		}
		msg.Blocks = append(msg.Blocks, block)
	default:
		block, err := cmd.ameshNow(ctx, now)
		if err != nil {
			return err
		}
		msg.Blocks = append(msg.Blocks, block)
	}
	_, err = client.PostMessage(ctx, msg)
	return err

}

func (cmd AmeshCommand) ameshNow(ctx context.Context, now time.Time) (block slack.Block, err error) {

	entry := amesh.GetEntry(now)

	bname := os.Getenv("GOOGLE_STORAGE_BUCKET_NAME")
	datetime := entry.Time.Format("2006-0102-1504")
	fname := fmt.Sprintf("%s.png", datetime)
	furl := cmd.Storage.URL(bname, fname)

	exists, err := cmd.Storage.Exists(ctx, bname, fname)
	if err != nil {
		return nil, err
	}
	if exists {
		return slack.NewImageBlock(furl, datetime, "", nil), nil
	}

	if _, err = entry.GetImage(true, true); err != nil {
		return nil, fmt.Errorf("failed to get image of amesh: %v", err)
	}

	if err := cmd.uploadEntryToStorage(ctx, entry, bname); err != nil {
		return nil, err
	}

	return slack.NewImageBlock(furl, datetime, "", nil), nil
}

func (cmd AmeshCommand) ameshAnimated(ctx context.Context, now time.Time) (block slack.Block, err error) {

	entries := amesh.GetEntries(now.Add(-40*time.Minute), now)

	bname := os.Getenv("GOOGLE_STORAGE_BUCKET_NAME")
	datetime := entries[0].Time.Format("2006-0102-1504")
	fname := fmt.Sprintf("%s.gif", datetime)
	furl := cmd.Storage.URL(bname, fname)

	exists, err := cmd.Storage.Exists(ctx, bname, fname)
	if err != nil {
		return nil, err
	}
	if exists {
		return slack.NewImageBlock(furl, datetime, "", nil), nil
	}

	g, err := entries.ToGif(500, true)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err := gif.EncodeAll(buf, g); err != nil {
		return nil, err
	}

	if err := cmd.Storage.Upload(ctx, bname, fname, buf.Bytes()); err != nil {
		return nil, err
	}

	go cmd.uploadEntriesToStorage(context.Background(), entries, bname)

	return slack.NewImageBlock(furl, datetime, "", nil), nil
}

func (cmd AmeshCommand) uploadEntriesToStorage(ctx context.Context, entries amesh.Entries, bucket string) error {
	for _, entry := range entries {
		if err := cmd.uploadEntryToStorage(ctx, entry, bucket); err != nil {
			fmt.Printf("[ERROR] %v.uploadEntriesToStorage: %v\n", cmd, err)
			return err
		}
	}
	return nil
}

func (cmd AmeshCommand) uploadEntryToStorage(ctx context.Context, entry *amesh.Entry, bucket string) error {
	if entry.Image == nil {
		return fmt.Errorf("image for the entry:%v is nil, use GetImage first", entry.Time)
	}
	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, entry.Image); err != nil {
		return err
	}
	efname := fmt.Sprintf("%s.png", entry.Time.Format("2006-0102-1504"))
	if err := cmd.Storage.Upload(ctx, bucket, efname, buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// Help ヘルプ一覧に表示されるもの
func (cmd AmeshCommand) Help() string {
	return "アメッシュ表示コマンド\n```@amesh [-a] [-h]```"
}
