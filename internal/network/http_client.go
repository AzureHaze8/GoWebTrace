package network

import (
	"GoWebTrace/pkg"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 发送 HTTP 请求，并返回包含Server和X-Powered-By响应头的map
func SendRequest(urls string, extractTLS bool, proxy string) (*ResponseInfo, error) {
	var redirectHistory []string

	// cert证书配置
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,             // 忽略TLS证书验证，允许连接自签名或无效证书的设备
			MinVersion:         tls.VersionTLS10, // 确保对老旧协议和加密套件的兼容性
			MaxVersion:         tls.VersionTLS13,
			CipherSuites:       DefaultCipherSuites, // 使用默认的加密套件
		},
	}
	// 是否存在代理
	if err := pkg.ConfigureTransport(transport, proxy); err != nil {
		return nil, err
	}

	// 创建一个可配置的HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second, // 10秒超时
		// 为了捕获跳转逻辑，我们自定义CheckRedirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 将上一个请求的URL记录到历史中
			if len(via) > 0 {
				redirectHistory = append(redirectHistory, via[len(via)-1].URL.String())
			}
			// 允许最多3次重定向
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
		Transport: transport,
	}

	// 创建请求
	req, err := http.NewRequest("GET", urls, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%w", err)
	}

	// 设置随机的User-Agent
	req.Header.Set("User-Agent", pkg.GetRandomUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "close")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%w", err)
	}
	defer resp.Body.Close()

	// 读取响应正文
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应正文失败：%w", err)
	}

	// 提取 TLS 指纹信息
	var tlsInfo *TLSInfo
	if extractTLS && resp.TLS != nil {
		tlsInfo = ExtractTLSInfo(resp.TLS)
	}

	// 填充结构体
	finalURL := resp.Request.URL
	info := &ResponseInfo{
		URL:             finalURL.String(), // 已获取 URL
		Path:            finalURL.Path,     // 已获取 Path
		StatusCode:      resp.StatusCode,   // 已获取 StatusCode
		Headers:         resp.Header,       // 已获取 Headers
		Body:            body,              // 已获取 Body
		Cookies:         resp.Cookies(),    // 已获取 Cookies
		RedirectHistory: redirectHistory,   // 已获取 RedirectHistory
		TLS:             tlsInfo,           // 已获取 TLS
	}
	return info, nil
}
