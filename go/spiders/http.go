package spiders

import (
	"bytes"
	"crypto/tls"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/u"
)

type HttpSpider struct {
	charset string
	client  *http.Client
}

type HttpSpiderOptions struct {
	// 可选
	Charset string
	// 可选
	ProxyURL string
	// 可选
	Client *http.Client
}

func NewHttpSpider(options *HttpSpiderOptions) *HttpSpider {
	r := new(HttpSpider)

	if options != nil {
		r.charset = options.Charset
		r.client = options.Client
	}

	if r.client == nil {
		// 使用默认Client
		r.client = http.DefaultClient
	}

	if options != nil && options.ProxyURL != "" {
		// 设置代理
		r.client.Transport = &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return url.Parse(options.ProxyURL)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return r
}

func (self *HttpSpider) Get(targetURL string) (*goquery.Document, error) {
	return self.send("GET", targetURL, "")
}

func (self *HttpSpider) Post(targetURL, body string) (*goquery.Document, error) {
	return self.send("POST", targetURL, body)
}

func (self *HttpSpider) send(method, targetURL, bodyStr string) (*goquery.Document, error) {
	userAgent := UserAgents[rand.Intn(len(UserAgents))]

	var body io.Reader
	if bodyStr != "" {
		body = bytes.NewReader(u.StrToBytes(bodyStr))
	}

	req, err := http.NewRequest(method, targetURL, body)
	if err != nil {
		return nil, serr.WithStack(err)
	}
	req.Header = http.Header{"User-Agent": []string{userAgent}}
	if method == http.MethodPost {
		req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	}

	resp, err := self.client.Do(req)
	if err != nil {
		return nil, serr.WithStack(err)
	}
	defer resp.Body.Close()

	var htmlReader io.Reader
	htmlReader = resp.Body
	if self.charset != "" {
		htmlReader, err = decode(resp.Body, self.charset)
		if err != nil {
			return nil, serr.WithStack(err)
		}
	}

	doc, err := goquery.NewDocumentFromReader(htmlReader)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	return doc, nil
}

func (self *HttpSpider) ExecuteRules(data map[string]interface{}, rules []interface{}) error {
	for _, rule := range rules {
		for k, v := range rule.(map[string]interface{}) {
			switch k {
			case "GET":
				value := getValue(data, v)
				array := strings.Split(value, "->")
				doc, err := self.Get(array[0])
				if err != nil {
					return err
				}

				toKey := getDataKey(array[1])
				data[toKey] = doc
				break
			case "TEXT":
				value := getValue(data, v)
				array := strings.Split(value, "->")
				if len(array) != 3 {
					break
				}

				fromKey := getDataKey(array[0])
				query := data[fromKey].(*goquery.Selection)

				nodeText := GetHttpText(query, array[0])

				toKey := getDataKey(array[1])
				data[toKey] = nodeText

				break
			case "LIST":
				subRules := v.(map[string]interface{})
				fromKey := getDataKey(subRules["From"].(string))
				doc := data[fromKey].(*goquery.Document)
				selector := subRules["Selector"].(string)
				each := subRules["Each"].([]interface{})

				items := make([]map[string]interface{}, 0)
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					item := map[string]interface{}{"node": s}

					err := self.ExecuteRules(item, each)
					if u.LogError(err) {
						return
					}

					items = append(items, item)
				})

				toKey := getDataKey(subRules["To"].(string))
				data[toKey] = items
				break
			}
		}
	}

	return nil
}
