package spiders

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
	"github.com/syncfuture/go/u"
)

type BrowserSpider struct {
	ProxyURL    *url.URL
	Context     context.Context
	CancelFuncs []context.CancelFunc
}

type BrowserSpiderOptions struct {
	RemoteURL string
	Proxy     string
	Headless  bool
	Incognito bool
}

func NewBrowserSpider(ctx context.Context, inOptions *BrowserSpiderOptions) *BrowserSpider {
	if inOptions == nil {
		inOptions = &BrowserSpiderOptions{
			Headless:  false,
			Incognito: true,
		}
	}

	r := new(BrowserSpider)
	r.CancelFuncs = make([]context.CancelFunc, 0)

	opts := append(chromedp.DefaultExecAllocatorOptions[:])

	if !inOptions.Headless {
		opts = append(opts,
			chromedp.Flag("headless", false),
			chromedp.Flag("hide-scrollbars", false),
			chromedp.Flag("mute-audio", false),
			chromedp.Flag("incognito", true),
		)
	}

	if inOptions.Incognito {
		opts = append(opts,
			chromedp.Flag("incognito", true),
		)
	}

	if inOptions.Proxy != "" {
		var err error
		r.ProxyURL, err = url.Parse(inOptions.Proxy)
		if err != nil {
			u.LogError(err)
		}

		opts = append(opts, chromedp.ProxyServer(fmt.Sprintf("%s://%s", r.ProxyURL.Scheme, r.ProxyURL.Host)))
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	r.CancelFuncs = append(r.CancelFuncs, cancel)
	if inOptions.RemoteURL == "" {
		ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
		r.CancelFuncs = append(r.CancelFuncs, cancel)
	} else {
		ctx, cancel = chromedp.NewRemoteAllocator(ctx, inOptions.RemoteURL)
		r.CancelFuncs = append(r.CancelFuncs, cancel)
	}

	ctx, cancel = chromedp.NewContext(ctx)
	r.CancelFuncs = append(r.CancelFuncs, cancel)
	r.Context = ctx

	if r.ProxyURL != nil && r.ProxyURL.User != nil {
		pass, _ := r.ProxyURL.User.Password()

		chromedp.ListenTarget(ctx, func(ev interface{}) {
			go func() {
				switch ev := ev.(type) {
				case *fetch.EventAuthRequired:
					// 弹出代理登录后自动填写用户名密码
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)

					resp := &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: r.ProxyURL.User.Username(),
						Password: pass,
					}

					err := fetch.ContinueWithAuth(ev.RequestID, resp).Do(execCtx)
					if err != nil {
						log.Print(err)
					}

				case *fetch.EventRequestPaused:
					c := chromedp.FromContext(ctx)
					execCtx := cdp.WithExecutor(ctx, c.Target)
					err := fetch.ContinueRequest(ev.RequestID).Do(execCtx)
					if err != nil {
						log.Print(err)
					}
				}
			}()
		})
	}

	return r
}

func (self *BrowserSpider) Cancels() {
	for i := range self.CancelFuncs {
		self.CancelFuncs[len(self.CancelFuncs)-1-i]()
	}
}

func (self *BrowserSpider) Run(actions ...chromedp.Action) error {
	if self.ProxyURL != nil && self.ProxyURL.User != nil {
		// 代理需要登录，必须添加WithHandleAuthRequests(true)
		actions = append([]chromedp.Action{fetch.Enable().WithHandleAuthRequests(true)}, actions...)
	}

	err := chromedp.Run(self.Context, actions...)
	return err
}
