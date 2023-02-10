package spiders

import (
	"context"
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// 使用代理例子
func TestProxy(t *testing.T) {
	spider := NewBrowserSpider(context.Background(), &BrowserSpiderOptions{
		Proxy:     "http://cjtvtbtz-rotate:51bxj0ldmvdc@p.webshare.io",
		Headless:  false,
		Incognito: true,
	})
	// defer spider.Cancels()

	var html string
	err := spider.Run(
		chromedp.Navigate("https://checker.soax.com/api/ipinfo"),
		chromedp.OuterHTML("html", &html),
	)

	require.NoError(t, err)

	fmt.Println(html)
}
