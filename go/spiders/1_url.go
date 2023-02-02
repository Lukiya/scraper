package spiders

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/syncfuture/go/u"
)

func AbsURL(baseURL string, urls ...string) string {
	if len(urls) == 0 || urls[0] == "" {
		// 没有传入额外地址，直接返回基地址
		return baseURL
	}

	if strings.HasPrefix(urls[0], "http://") || strings.HasPrefix(urls[0], "https://") {
		// 已经是绝对地址，直接返回
		if len(urls) == 1 {
			return urls[0]
		} else {
			return JoinURLs(urls[0], urls[1:]...)
		}
	}

	if strings.HasPrefix(urls[0], "/") {
		// /开头，表示相对于根地址
		baseUri, err := url.Parse(baseURL)
		if u.LogError(err) {
			return ""
		}

		rootURL := fmt.Sprintf("%s://%s", baseUri.Scheme, baseUri.Host)
		return JoinURLs(rootURL, urls...)
	} else {
		// 否则表示相对于当前地址

		// 当前路径不能是文件
		hasFilename := strings.LastIndex(baseURL, ".") > strings.LastIndex(baseURL, "/")
		if hasFilename {
			baseURL, _ = filepath.Split(baseURL)
		}

		return JoinURLs(baseURL, urls...)
	}
}

func JoinURLs(baseURL string, urls ...string) string {
	r, err := url.JoinPath(baseURL, urls...)
	if u.LogError(err) {
		return ""
	}
	return r
}
