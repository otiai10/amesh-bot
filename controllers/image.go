package controllers

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strconv"

	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/disintegration/gift"
)

// Image ...
// botに対して、-filterが与えられたとき、botはSlack側に
// originではなく、このコントローラのエンドポイントを渡す.
// Slackの画像プロキシサーバ（具体的には https://slack-imgs.com/ ）は、
// originのURLとfilterのパラメータを含んだリクエストをここにGETするので、
// filter処理を施した画像バイナリをHTTPレスポンスとして返す.
func (c *Controller) Image(w http.ResponseWriter, req *http.Request) {

	origin := req.URL.Query().Get("url")
	filter := req.URL.Query().Get("filter")
	levelstr := req.URL.Query().Get("level")

	// {{{ キャッシュ画像存在確認
	cachekey := req.URL.Query().Encode()
	// FIXME: これはDI的にStorageから提供されるべきでは？
	bname := os.Getenv("GOOGLE_STORAGE_BUCKET_NAME")
	ctx := req.Context()
	exists, err := c.Storage.Exists(ctx, bname, cachekey)
	if err == nil && exists {
		rc, err := c.Storage.Get(ctx, bname, cachekey)
		if err == nil && rc != nil {
			if _, err := io.Copy(w, rc); err != nil {
				fmt.Println("[ERROR] io.Copy", err.Error())
			} else if err := rc.Close(); err != nil {
				fmt.Println("[ERROR] obj.Close", err.Error())
			}
			return
		}
	}
	// }}}

	level := 60
	if lv, err := strconv.Atoi(levelstr); err == nil {
		level = lv
	}

	// originalの画像バイナリを取得.
	res, err := http.Get(origin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		w.WriteHeader(res.StatusCode)
		return
	}

	src, fmtname, err := image.Decode(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	g := gift.New()
	switch filter {
	default:
		g.Add(gift.Pixelate(level))
	}

	dest := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dest, src)

	// このエンドポイントへGETをかけるクライアントは
	// 標準的なブラウザではなくて, slack-imgs.com なので,
	// ここでCache-Controlを返しても意味は無かった.
	w.Header().Add("Cache-Control", "public,max-age=3600,immutable")

	switch fmtname {
	case "png":
		w.Header().Add("Content-Type", "image/png")
		png.Encode(w, dest)
	case "gif":
		w.Header().Add("Content-Type", "image/gif")
		gif.Encode(w, dest, nil)
	case "jpeg":
		w.Header().Add("Content-Type", "image/jpeg")
		jpeg.Encode(w, dest, nil)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	// {{{ 以下、キャッシュとしてCloudStorageへ保存
	go c.cacheFilteredImage(context.Background(), bname, cachekey, fmtname, dest)
	// }}}
}

func (c *Controller) cacheFilteredImage(ctx context.Context, bname, cachekey, fmtname string, img image.Image) (err error) {
	buf := bytes.NewBuffer(nil)
	switch fmtname {
	case "png":
		err = png.Encode(buf, img)
	case "gif":
		err = gif.Encode(buf, img, nil)
	case "jpeg":
		err = jpeg.Encode(buf, img, nil)
	}
	if err != nil {
		fmt.Printf("[ERROR] c.Image::cacheFilteredImage::encode %v", err)
		return err
	}
	if err := c.Storage.Upload(ctx, bname, cachekey, buf.Bytes()); err != nil {
		fmt.Printf("[ERROR] c.Image::cacheFilteredImage::upload %v", err)
		return err
	}
	return nil

}
