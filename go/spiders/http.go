package spiders

import (
	"bytes"
	"crypto/tls"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/u"
	"golang.org/x/text/encoding/htmlindex"
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

func decode(reader io.Reader, charset string) (io.Reader, error) {
	r := reader

	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, serr.WithStack(err)
	}

	if name, _ := htmlindex.Name(e); name != "utf-8" {
		r = e.NewDecoder().Reader(reader)
	}

	return r, nil
}
