package output

import (
	"time"
)

// 解析截图 API 返回的 JSON 数据
type apiResponse struct {
	Status int `json:"status"`
	Data   struct {
		Thumb struct {
			URL string `json:"url"`
		} `json:"thumb"`
	} `json:"data"`
}

// 存储扫描结果
type Result struct {
	ID             int
	URL            string
	StatusCode     int
	ContentLength  int
	Title          string
	CMS            string
	ScreenshotPath string
	Timestamp      time.Time
}

// 创建一个新的 Result 实例
func NewResult(id int, url string, statusCode int, contentLength int, title string, cms string, screenshotPath string) *Result {
	return &Result{
		ID:             id,
		URL:            url,
		StatusCode:     statusCode,
		ContentLength:  contentLength,
		Title:          title,
		CMS:            cms,
		ScreenshotPath: screenshotPath,
		Timestamp:      time.Now(),
	}
}