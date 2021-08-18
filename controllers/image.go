package controllers

import (
	"image"
	"net/http"
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

}
