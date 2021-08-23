package commands

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/otiai10/goapis/google"
)

type SearchIntent struct {
	url.Values
	Retry   int
	Request google.CustomSearchRequest
}

func RecoverIntent(ctx context.Context) *SearchIntent {
	intent, ok := ctx.Value(imageSearchIntentCtxKey).(*SearchIntent)
	if ok {
		intent.Values = url.Values{}
		return intent
	}
	return &SearchIntent{Values: url.Values{}}
}

func (intent *SearchIntent) Unsafe(unsafe bool) {
	if unsafe {
		intent.Add("safe", "off")
	} else {
		intent.Add("safe", "active")
	}
}

func (intent *SearchIntent) Build() url.Values {
	if intent.Request.StartIndex == 0 {
		intent.Values.Add("start", fmt.Sprintf("%d", rand.Intn(90)))
	} else {
		start := intent.Request.StartIndex - 10
		if start < 0 {
			start = 0
		}
		intent.Values.Add("start", fmt.Sprintf("%d", start))
	}
	return intent.Values
}
