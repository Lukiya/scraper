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
func TestBrowserSearchBooks(t *testing.T) {
	config, err := rules.ReadRulesFromFile("rules_sample.json")
	require.NoError(t, err)
	require.NotEmpty(t, config)

	data := config["Data"].(map[string]interface{})
	searchBook := config["SearchBooks"].(map[string]interface{})
	searchBookMode := searchBook["Mode"].(string)

	if searchBookMode == "Browser" {
		spider := NewBrowserSpider(context.Background(), &BrowserSpiderOptions{
			Proxy:     "http://cjtvtbtz-rotate:51bxj0ldmvdc@p.webshare.io",
			Headless:  false,
			Incognito: true,
			Timeout:   30 * time.Second,
		})
		defer spider.Cancel()

		rules := searchBook["Rules"].([]interface{})
		data["BookName"] = "灵境行者"

		err = spider.ExecuteRules(data, rules, nil)
		require.NoError(t, err)
		require.NotEmpty(t, data["Books"])
	}
}

// 执行规则
func TestBrowserGetChatpers(t *testing.T) {
	config, err := rules.ReadRulesFromFile("rules_sample.json")
	require.NoError(t, err)
	require.NotEmpty(t, config)

	data := config["Data"].(map[string]interface{})
	searchBook := config["GetChapters"].(map[string]interface{})
	searchBookMode := searchBook["Mode"].(string)

	if searchBookMode == "Browser" {
		spider := NewBrowserSpider(context.Background(), &BrowserSpiderOptions{
			// Proxy:     "http://cjtvtbtz-rotate:51bxj0ldmvdc@p.webshare.io",
			Headless:  false,
			Incognito: true,
			Timeout:   3600 * time.Second,
		})
		defer spider.Cancel()

		rules := searchBook["Rules"].([]interface{})

		err = spider.ExecuteRules(data, rules, nil)
		require.NoError(t, err)
		require.NotEmpty(t, data["ChapterPages"])
	}
}
