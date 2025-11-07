package network

import (
	"GoWebTrace/pkg"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/twmb/murmur3"
	"golang.org/x/net/html"
)

// 解析HTML and extract links, scripts, title, css, js, favicon
func (info *ResponseInfo) ParseHTML(proxy string) {
	doc, err := html.Parse(strings.NewReader(string(info.Body)))
	if err != nil {
		return
	}

	// 获取基础URL，用于解析相对URL
	base, err := url.Parse(info.URL)
	if err != nil {
		return // 如果基础URL无效，就不能解析相对URL
	}

	// 遍历HTML树，提取信息
	var f func(*html.Node)
	f = func(n *html.Node) {
		// 提取标题
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				info.Title = n.FirstChild.Data
			}
		}

		// 提取JS文件
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					absURL := toAbsoluteURL(base, a.Val)
					if absURL != "" {
						info.JsFiles = append(info.JsFiles, absURL)
					}
				}
			}
		}

		// 提取CSS文件和Favicon
		if n.Type == html.ElementNode && n.Data == "link" {
			isCSS := false
			isFavicon := false
			href := ""
			for _, a := range n.Attr {
				if a.Key == "rel" && (a.Val == "stylesheet" || a.Val == "preload" && n.Data == "style") {
					isCSS = true
				}
				if a.Key == "rel" && (a.Val == "icon" || a.Val == "shortcut icon") {
					isFavicon = true
				}
				if a.Key == "href" {
					href = a.Val
				}
			}
			if href != "" {
				absURL := toAbsoluteURL(base, href)
				if absURL != "" {
					if isCSS {
						info.CssFiles = append(info.CssFiles, absURL)
					}
					if isFavicon {
						info.Favicon = absURL
					}
				}
			}
		}

		// 循环访问所有子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	if info.Favicon == "" && base != nil {
		info.Favicon = base.Scheme + "://" + base.Host + "/favicon.ico"
	}

	//在确定了最终的Favicon URL后，填充ResponseInfo结构体
	if info.Favicon != "" {
		// 调用函数获取 mmh3 和 md5 哈希值
		mmh3Hash, md5Hash, err := GetFaviconHash(info.Favicon, proxy)
		if err == nil {
			// 如果获取成功，则将哈希值赋给info对象的相应字段
			info.FaviconMmh3 = mmh3Hash
			info.FaviconMd5 = md5Hash
		}
	}
}

// 基准URL转换为绝对URL
func toAbsoluteURL(base *url.URL, href string) string {
	// 优先处理绝对URL
	if strings.HasPrefix(href, "//") {
		// 处理协议相对URL
		return base.Scheme + ":" + href
	}

	// 处理所有其他类型的URL
	u, err := url.Parse(href)
	if err != nil {
		return "" // 解析失败，返回空字符串
	}
	return base.ResolveReference(u).String()
}

// 下载Favicon并计算其mmh3哈希值和md5哈希值
func GetFaviconHash(faviconURL string, proxy string) (string, string, error) {
	// 忽略TLS证书错误
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// 如果提供了代理，则设置代理
	if err := pkg.ConfigureTransport(transport, proxy); err != nil {
		return "", "", err
	}

	// 创建一个HTTP客户端，配置为忽略TLS证书错误，并设置10秒超时
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	// 创建一个GET请求
	req, err := http.NewRequest("GET", faviconURL, nil)
	if err != nil {
		return "", "", err
	}
	// 设置随机的User-Agent
	req.Header.Set("User-Agent", pkg.GetRandomUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Connection", "close")

	// 发送GET请求下载 Favicon
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err // 如果下载失败，返回空字符串
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err // 读取失败，返回空字符串
	}
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", "", err // 如果状态码不是200 OK，说明获取失败
	}

	// 计算mmh3哈希值
	b64body := base64.StdEncoding.EncodeToString(body)
	hasher := murmur3.New32()
	_, _ = hasher.Write([]byte(b64body)) // 对响应体进行Base64编码， 计算mmh3哈希
	mmh3Hash := hasher.Sum32()

	// 计算MD5哈希
	md5Hasher := md5.New()
	_, _ = md5Hasher.Write(body) // MD5直接计算原始字节
	md5Hash := hex.EncodeToString(md5Hasher.Sum(nil))

	// 返回结果
	return strconv.Itoa(int(mmh3Hash)), md5Hash, nil
}

// 将HTTP响应头转换为单个字符串，每个头字段以 "Key: Value" 的形式表示，并用换行符分隔
func (r *ResponseInfo) GetHeadersAsString() string {
	var headers strings.Builder
	for key, values := range r.Headers {
		for _, value := range values {
			headers.WriteString(key)
			headers.WriteString(": ")
			headers.WriteString(value)
			headers.WriteString("\n")
		}
	}
	return headers.String()
}
