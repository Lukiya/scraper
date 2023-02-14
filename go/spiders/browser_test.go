package spiders

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Lukiya/scraper/go/rules"
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

// 执行规则
func TestExecuteRules(t *testing.T) {
	a, err := rules.ReadRulesFromFile("rules_sample.json")
	require.NoError(t, err)
	require.NotEmpty(t, a)

	spider := NewBrowserSpider(context.Background(), &BrowserSpiderOptions{
		Proxy:     "http://cjtvtbtz-rotate:51bxj0ldmvdc@p.webshare.io",
		Headless:  false,
		Incognito: true,
		Timeout:   3600 * time.Second,
	})
	defer spider.Cancel()

	data := a["data"].(map[string]interface{})
	rules := a["BrowserRules"].([]interface{})
	data["BookName"] = "灵境行者"

	err = spider.ExecuteRules(data, rules, nil)
	require.NoError(t, err)
	require.NotEmpty(t, data["Books"])
}
