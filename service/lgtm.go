package service

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type (
	LGTM struct{}

	// <html><body>You are being <a href="https://lgtm.lol/i/244?random">redirected</a>.</body></html>%
	LGTMResponseHTML struct {
		Anchor struct {
			Href string `xml:"href,attr"`
		} `xml:"body>a"`
	}
	// [![LGTM](https://lgtm.lol/p/244)](https://lgtm.lol/i/244)
)

func (lgtm LGTM) Random() (string, string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get("https://lgtm.lol/random")
	if err != nil {
		return "", "", fmt.Errorf("http error: %v", err)
	}
	defer res.Body.Close()
	payload := LGTMResponseHTML{}
	if err := xml.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %v", err.Error())
	}
	u, err := url.Parse(payload.Anchor.Href)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL format given: %v", err.Error())
	}
	id := strings.Split(u.Path, "/")[2]
	imgurl := fmt.Sprintf("%s://%s/p/%s", u.Scheme, u.Host, id)
	mrkdwn := fmt.Sprintf("[![LGTM](%[1]s://%[2]s/p/%[3]s)](%[1]s://%[2]s/i/%[3]s)", u.Scheme, u.Host, id)
	return imgurl, mrkdwn, nil
}
