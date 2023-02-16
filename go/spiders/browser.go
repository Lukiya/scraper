package spiders

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
	"github.com/syncfuture/go/serr"
	"github.com/syncfuture/go/slog"
	"github.com/syncfuture/go/u"
)

type BrowserSpider struct {
	ProxyURL    *url.URL
	Context     context.Context
	CancelFuncs []context.CancelFunc
}

type BrowserSpiderOptions struct {
	RemoteURL  string
	Proxy      string
	WindowSize string
	Headless   bool
	Incognito  bool
	Timeout    time.Duration
}

func NewBrowserSpider(ctx context.Context, inOptions *BrowserSpiderOptions) *BrowserSpider {
	if inOptions == nil {
		inOptions = &BrowserSpiderOptions{
			Headless:  false,
			Incognito: true,
		}
	}

	if inOptions.Timeout == 0 {
		inOptions.Timeout = 30 * time.Second
	}

	if inOptions.WindowSize == "" {
		inOptions.WindowSize = "1280,900"
	}

	r := new(BrowserSpider)
	r.CancelFuncs = make([]context.CancelFunc, 0)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("window-size", inOptions.WindowSize),
	)

	if !inOptions.Headless {
		opts = append(opts,
			chromedp.Flag("headless", false),
			chromedp.Flag("hide-scrollbars", false),
			chromedp.Flag("mute-audio", false),
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

	ctx, cancel := context.WithTimeout(ctx, inOptions.Timeout)
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

func (self *BrowserSpider) Cancel() {
	for i := range self.CancelFuncs {
		self.CancelFuncs[len(self.CancelFuncs)-1-i]()
	}
}

func (self *BrowserSpider) Run(actions ...chromedp.Action) error {
	// 判断是否有导航动作
	var hasNavi bool
	for _, x := range actions {
		if _, ok := x.(chromedp.NavigateAction); ok {
			hasNavi = true
			break
		}
	}

	// 有导航并有代理，且代理需要认证，必须添加WithHandleAuthRequests(true)
	if hasNavi && self.ProxyURL != nil && self.ProxyURL.User != nil {
		actions = append([]chromedp.Action{fetch.Enable().WithHandleAuthRequests(true)}, actions...)
	}

	err := chromedp.Run(self.Context, actions...)
	if err != nil {
		return serr.WithStack(err)
	}

	return nil
}

func (self *BrowserSpider) ExecuteRules(data map[string]interface{}, rules []interface{}, node *cdp.Node) error {
	opts := make([]chromedp.QueryOption, 0)
	// if n, ok := data["node"]; ok {
	// 	opts = append(opts, chromedp.FromNode(n.(*cdp.Node)))
	// }

	for _, rule := range rules {
		for k, v := range rule.(map[string]interface{}) {
			switch k {
			case "NAVI":
				value := getValue(data, v)
				err := self.Run(chromedp.Navigate(value))
				if err != nil {
					a, _ := json.Marshal(data)
					slog.Error("#################", value, u.BytesToStr(a))
					return err
				}

				break
			case "SETVAL":
				value := getValue(data, v)
				array := strings.Split(value, "->")
				// 从context里取数据
				contextKey := getDataKey(array[0])
				contextValue := data[contextKey].(string)
				// 将contextValue放入指定selector
				err := chromedp.Run(self.Context, chromedp.SetValue(array[1], contextValue, opts...))
				if err != nil {
					return err
				}
				break
			case "CLICK":
				value := getValue(data, v)
				err := chromedp.Run(self.Context, chromedp.Click(value, opts...))
				if err != nil {
					return err
				}
				break
			case "TEXT":
				value := getValue(data, v)
				array := strings.Split(value, "->")

				var nodeText string
				err := chromedp.Run(self.Context, self.GetText(array[0], &nodeText, node, opts...))
				if err != nil {
					return err
				}

				if nodeText == "" {
					print(array[1])
				}

				toKey := getDataKey(array[1])
				data[toKey] = CleanText(nodeText)
				break
			case "LIST":
				subRules := v.(map[string]interface{})
				selector := subRules["Selector"].(string)
				each := subRules["Each"].([]interface{})

				var nodes []*cdp.Node
				opts = append(opts, chromedp.ByQueryAll)
				err := chromedp.Run(self.Context, chromedp.Nodes(selector, &nodes, opts...))
				if err != nil {
					return err
				}

				items := make([]map[string]interface{}, 0)

				for _, node := range nodes {
					item := map[string]interface{}{}
					err := self.ExecuteRules(item, each, node)
					if err != nil {
						return err
					}

					items = append(items, item)
				}

				toKey := getDataKey(subRules["To"].(string))
				data[toKey] = items
				break
			}
		}
	}

	return nil
}

func (self *BrowserSpider) GetText(sel string, out *string, node *cdp.Node, opts ...chromedp.QueryOption) chromedp.Action {
	array := strings.Split(sel, "@")
	if len(array) == 2 {
		if node != nil && array[0] == "" {
			return chromedp.ActionFunc(func(ctx context.Context) error {
				*out = node.AttributeValue(array[1])
				return nil
			})
		}

		var ok bool
		if node != nil {
			opts = append(opts, chromedp.FromNode(node))
		}
		return chromedp.AttributeValue(array[0], array[1], out, &ok, opts...)
	}

	action := chromedp.Text(sel, out, opts...)
	return action
}
