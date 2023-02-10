package spiders

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:9222/")
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ch := addNewTabListener(ctx)
	// var htmlContent string
	// var imgBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.xhwx3.com/"),
		// chromedp.Sleep(5*time.Second),
		// chromedp.FullScreenshot(&imgBuf, 100),
		// chromedp.SendKeys("#search input[name='kw']", "灵境行者"),
		chromedp.SetValue("#search input[name='kw']", "灵境行者"),
		chromedp.Click("#search_btn"),
	)

	require.NoError(t, err)

	newCtx, cancel := chromedp.NewContext(ctx, chromedp.WithTargetID(<-ch))
	defer cancel()

	var res string
	err = chromedp.Run(newCtx,
		chromedp.OuterHTML(`html`, &res, chromedp.BySearch),
	)

	print(res)
}

func addNewTabListener(ctx context.Context) <-chan target.ID {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	return chromedp.WaitNewTarget(ctx, func(info *target.Info) bool {
		return info.URL != ""
	})
}
