package pkg

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/fatih/color"
	"golang.org/x/net/proxy"
)

var proxyListPath = "config/proxyList.txt"

// 优先从 proxyList 加载代理列表，如果路径为空，则尝试使用 proxyURL
func New(proxyList bool, proxyURL string) (*Manager, error) {
	var proxies []string

	if proxyList {
		if _, err := os.Stat(proxyListPath); err == nil {
			proxies, err = ReadLines(proxyListPath)
			if err != nil {
				return nil, err
			}
			log.Printf(color.GreenString("[*]从 %s 加载了 %d 个代理\n"), proxyListPath, len(proxies))
		} else {
			log.Printf(color.YellowString("[-]警告: 无法找到代理列表文件 %s\n"), proxyListPath)
		}
	} else if proxyURL != "" {
		proxies = []string{proxyURL}
		log.Printf(color.GreenString("[*]使用了单个代理: %s\n"), proxyURL)
	}

	return &Manager{proxies: proxies}, nil
}

// 轮询代理，没有可用的代理，返回空字符串
func (m *Manager) GetProxy() string {
	if len(m.proxies) == 0 {
		return ""
	}
	idx := atomic.AddUint64(&m.count, 1) - 1
	return m.proxies[int(idx%uint64(len(m.proxies)))]
}

// 检查管理器中是否有代理
func (m *Manager) HasProxies() bool {
	return len(m.proxies) > 0
}

// 根据代理字符串配置 http.Transport
func ConfigureTransport(transport *http.Transport, proxyString string) error {
	if proxyString == "" {
		return nil
	}

	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		log.Printf(color.RedString("[!]无效的代理URL: %s\n"), proxyString)
	}

	switch proxyURL.Scheme {
	case "http", "https":
		transport.Proxy = http.ProxyURL(proxyURL)
	case "socks4", "socks5":
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			log.Printf(color.RedString("[!]创建SOCKS拨号器失败: %v\n"), err)
		}
		contextDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			log.Println(color.RedString("[!]SOCKS拨号器不支持上下文 DialContext"))
		}
		transport.DialContext = contextDialer.DialContext
	default:
		log.Printf(color.RedString("[!]不支持的代理协议: %s\n"), proxyURL.Scheme)
	}
	return nil
}
