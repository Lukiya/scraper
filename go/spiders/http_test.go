package spiders

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/Lukiya/scraper/go/rules"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	httpSpider := NewHttpSpider(&HttpSpiderOptions{
		ProxyURL: "http://cjtvtbtz-rotate:51bxj0ldmvdc@p.webshare.io",
	})

	doc, err := httpSpider.Get("https://www.xhwx3.com")
	require.NoError(t, err)

	token := GetHttpText(doc.Children(), "input[name='_token']@value")

	body := fmt.Sprintf("_token=%s&kw=%s", token, url.QueryEscape("灵境行者"))
	doc, err = httpSpider.Post("https://www.xhwx3.com/search", body)
	require.NoError(t, err)

	// print(doc.Html())

	inputs := map[string]interface{}{
		"a": "div:nth-child(1)",
		"b": "div:nth-child(2)",
		"c": "div:nth-child(3)",
		"d": "div:nth-child(4)",
	}

	list := make([]map[string]interface{}, 0)

	doc.Find(".tableList li:nth-child(n+2)").Each(func(i int, s *goquery.Selection) {
		dic := make(map[string]interface{}, 0)
		for k, v := range inputs {
			dic[k] = GetHttpText(s, v.(string))
		}
		list = append(list, dic)
	})

}

func TestHttpExecuteRules(t *testing.T) {
	config, err := rules.ReadRulesFromFile("rules_sample.json")
	require.NoError(t, err)
	require.NotEmpty(t, config)

	data := config["Data"].(map[string]interface{})
	getChatpers := config["GetChapters"].(map[string]interface{})
	getChatpersMode := getChatpers["Mode"].(string)

	if getChatpersMode == "Http" {
		spider := NewHttpSpider(&HttpSpiderOptions{
			ProxyURL: "http://cjtvtbtz-rotate:51bxj0ldmvdc@p.webshare.io",
		})

		rules := getChatpers["Rules"].([]interface{})

		err = spider.ExecuteRules(data, rules)
		require.NoError(t, err)
		require.NotEmpty(t, data["Books"])
	}
}
