package output

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 请求在线截图并返回线上缩略图 URL
func SaveScreenshot(targetURL string) (string, error) {
	apiUrl := "https://nbapi.qikekeji.com/webshot/screenshot/"

	formData := url.Values{}
	formData.Set("url", targetURL)
	formData.Set("device", "desktop")
	formData.Set("type", "fullpage")
	formData.Set("width", "1280")
	formData.Set("height", "720")
	formData.Set("format", "jpg")
	formData.Set("quality", "100")

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("创建截图任务请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Referer", "https://webshot.cc/")
	req.Header.Set("Origin", "https://webshot.cc")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("执行截图任务请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("截图 API 返回非 200 状态码: %s", resp.Status)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("解码 API 响应失败: %w", err)
	}
	if apiResp.Status != 1 || apiResp.Data.Thumb.URL == "" {
		return "", fmt.Errorf("API 任务失败: %d", apiResp.Status)
	}

	onlineURL := apiResp.Data.Thumb.URL
	// log.Printf("获得线上截图地址: %s", onlineURL)
	return onlineURL, nil
}
