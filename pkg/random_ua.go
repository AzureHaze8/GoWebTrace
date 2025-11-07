package pkg

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
)

// ua.txt 中加载 User-Agent 字符串列表
var userAgents []string

// 默认ua
const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"

// 默认ua路径
var path = "config/ua.txt"

// init 加载 User-Agent
func init() {
	if err := loadUserAgents(path); err != nil {
		log.Printf("警告：无法从文件加载 User-Agent：%v。将使用默认 ua。", err)
	}
}

// 从文件读取，并填充 userAgents 切片
func loadUserAgents(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			userAgents = append(userAgents, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(userAgents) == 0 {
		log.Println("警告：ua.txt 为空或不包含有效的 User-Agent。")
	}

	return nil
}

// 从加载列表中选择并返回一个随机的 User-Agent 字符串
func GetRandomUserAgent() string {
	if len(userAgents) == 0 {
		// 如果列表为空，则返回默认的 User-Agent
		return defaultUserAgent
	}
	return userAgents[rand.Intn(len(userAgents))]
}
