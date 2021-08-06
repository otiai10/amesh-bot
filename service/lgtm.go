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
)

func (lgtm LGTM) Random() (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get("https://lgtm.lol/random")
	if err != nil {
		return "", fmt.Errorf("http error: %v", err)
	}
	defer res.Body.Close()
	payload := LGTMResponseHTML{}
	if err := xml.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err.Error())
	}
	u, err := url.Parse(payload.Anchor.Href)
	if err != nil {
		return "", fmt.Errorf("invalid URL format given: %v", err.Error())
	}
	return fmt.Sprintf("%s://%s/p/%s", u.Scheme, u.Host, strings.Split(u.Path, "/")[2]), nil
}
