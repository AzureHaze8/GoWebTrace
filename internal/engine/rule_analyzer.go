package engine

import (
	"GoWebTrace/internal/network"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)


func RuleAnalyzer(filePath string) (*network.LocationRules, error) {
	// 从文件加载规则
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析文件规则
	var db struct {
		Fingerprint []network.FingerprintRule `json:"fingerprint"`
	}
	err = json.Unmarshal(file, &db)
	if err != nil {
		return nil, fmt.Errorf("解析文件规则失败：%w", err)
	}

	// 转换关键字规则
	LocationRules := &network.LocationRules{}
	for _, rule := range db.Fingerprint {
		// 编译用于匹配MD5哈希的正则表达式
		md5Regex := regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

		switch rule.Location {
		case "body":
			// 如果位置是 "body"，再根据 "method" 进行第二层分类
			switch rule.Method {
			case "keyword":
				LocationRules.BodyKeywordRules = append(LocationRules.BodyKeywordRules, rule)
			case "regex":
				LocationRules.BodyRegexRules = append(LocationRules.BodyRegexRules, rule)
			case "faviconhash":
				// 智能判断当前哈希类型
				if md5Regex.MatchString(rule.Keyword[0]) {
					// 如果关键词匹配MD5格式 (32位十六进制字符串)
					LocationRules.FaviconMd5Hash = append(LocationRules.FaviconMd5Hash, rule)
				} else {
					// 否则，就是一个Murmur3哈希 (整数形式)
					LocationRules.FaviconMmh3Hash = append(LocationRules.FaviconMmh3Hash, rule)
				}
			}

		case "title":
			switch rule.Method {
			case "keyword":
				LocationRules.TitleKeywordRules = append(LocationRules.TitleKeywordRules, rule)
			case "regex":
				LocationRules.TitleRegexRules = append(LocationRules.TitleRegexRules, rule)
			}

		case "header":
			switch rule.Method {
			case "keyword":
				LocationRules.HeaderKeywordRules = append(LocationRules.HeaderKeywordRules, rule)
			case "regex":
				LocationRules.HeaderRegexRules = append(LocationRules.HeaderRegexRules, rule)
			}
		}
	}
	return LocationRules, nil
}
