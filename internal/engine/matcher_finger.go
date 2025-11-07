package engine

import (
	"GoWebTrace/internal/network"
	"log"
	"regexp"
	"strings"
)

// 对给定的响应信息匹配指纹库
func MatchFingerprints(info *network.ResponseInfo, rules *network.LocationRules) map[string]struct{} {
	matchedCMS := make(map[string]struct{})

	// Favicon Hash 匹配 (最高置信度)
	processHashRules(info.FaviconMmh3, rules.FaviconMmh3Hash, matchedCMS)
	processHashRules(info.FaviconMd5, rules.FaviconMd5Hash, matchedCMS)

	// Title规则匹配
	processRules(rules.TitleRegexRules, info.Title, matchedCMS, true)
	processRules(rules.TitleKeywordRules, info.Title, matchedCMS, false)

	// Header规则匹配
	var allHeaders strings.Builder
	allHeaders.WriteString(info.GetHeadersAsString())
	processRules(rules.HeaderRegexRules, allHeaders.String(), matchedCMS, true)
	processRules(rules.HeaderKeywordRules, allHeaders.String(), matchedCMS, false)

	// Body规则匹配
	// 先收集所有匹配规则及其版本信息
	bodyMatches := make(map[string]network.BodyMatch) // Key是基础CMS名
	processBodyRules(rules.BodyRegexRules, string(info.Body), bodyMatches, true)
	processBodyRules(rules.BodyKeywordRules, string(info.Body), bodyMatches, false)

	// 根据阈值判断Body匹配是否有效，并使用最佳名称更新最终结果
	const bodyMatchThreshold = 1 // Body部分至少需要1条规则命中才算有效
	for _, match := range bodyMatches {
		if match.Count >= bodyMatchThreshold {
			updateCMS(matchedCMS, match.Name)
		}
	}
	return matchedCMS
}

// 处理基于Favicon哈希值匹配
func processHashRules(hashValue string, hashRules []network.FingerprintRule, matchedCMS map[string]struct{}) {
	if hashValue == "" || len(hashRules) == 0 {
		return
	}
	for _, rule := range hashRules {
		if len(rule.Keyword) > 0 && rule.Keyword[0] == hashValue {
			updateCMS(matchedCMS, rule.CMS)
		}
	}
}

// 处理 Header 和 Title 规则
func processRules(rules []network.FingerprintRule, content string, matchedCMS map[string]struct{}, isRegex bool) {
	if content == "" || len(rules) == 0 {
		return
	}
	for _, rule := range rules {
		if matched, newFinger := checkRulesMatch(rule, content, isRegex); matched {
			updateCMS(matchedCMS, newFinger)
		}
	}
}

// 处理 Body 规则，收集匹配信息
func processBodyRules(rules []network.FingerprintRule, content string, bodyMatches map[string]network.BodyMatch, isRegex bool) {
	if content == "" || len(rules) == 0 {
		return
	}
	for _, rule := range rules {
		if matched, newFinger := checkBodyRulesMatch(rule, content, isRegex); matched {
			info := bodyMatches[rule.CMS]
			info.Count++
			// 优先使用带版本号的名称
			if newFinger != rule.CMS || info.Name == "" {
				info.Name = newFinger
			}
			bodyMatches[rule.CMS] = info
		}
	}
}

// 检查Header/Title规则，成功则返回(true, "带版本的名称")
func checkRulesMatch(rule network.FingerprintRule, content string, isRegex bool) (bool, string) {
	allMatch := true
	lowerContent := strings.ToLower(content)

	for _, keyword := range rule.Keyword {
		lowerKeyword := strings.ToLower(keyword)
		matchFound := false

		if isRegex {
			re, err := regexp.Compile(lowerKeyword)
			if err != nil {
				log.Printf("正则表达式编译错误: %v", err)
				continue
			}
			if re.MatchString(lowerContent) {
				matchFound = true
			}
		} else {
			if strings.Contains(lowerContent, lowerKeyword) {
				matchFound = true
			}
		}

		if !matchFound {
			allMatch = false
			break
		}
	}

	// 规则的所有关键字都匹配了，现在尝试提取版本
	if allMatch {
		versionedName := ExtractVersion(rule.CMS, content)
		if versionedName != "" {
			return true, versionedName
		}
		// 没找到版本，但规则匹配了，返回基础CMS名
		return true, rule.CMS
	}
	return false, ""
}

// 检查Body规则，成功则返回(true, "带版本的名称")
func checkBodyRulesMatch(rule network.FingerprintRule, content string, isRegex bool) (bool, string) {
	keywordCount := len(rule.Keyword)
	if keywordCount == 0 {
		return false, ""
	}

	matchCount := 0
	lowerContent := strings.ToLower(content)

	for _, keyword := range rule.Keyword {
		lowerKeyword := strings.ToLower(keyword)
		if isRegex {
			if re, err := regexp.Compile(lowerKeyword); err == nil && re.MatchString(lowerContent) {
				matchCount++
			}
		} else {
			if strings.Contains(lowerContent, lowerKeyword) {
				matchCount++
			}
		}
	}

	ruleMatched := false
	if keywordCount <= 2 {
		ruleMatched = (matchCount == keywordCount)
	} else {
		ruleMatched = (matchCount >= 2)
	}
	// 规则满足匹配阈值，现在尝试提取版本
	if ruleMatched {
		versionedName := ExtractVersion(rule.CMS, content)
		if versionedName != "" {
			return true, versionedName
		}
		return true, rule.CMS
	}

	return false, ""
}

// 从文本中提取产品名、版本号，如 "nginx1.18.0"
func ExtractVersion(productKeyword string, content string) string {
	lowerContent := strings.ToLower(content)
	lowerKeyword := strings.ToLower(productKeyword)
	escapedKeyword := regexp.QuoteMeta(lowerKeyword)

	re := regexp.MustCompile(
		escapedKeyword +
			`(?:[/\s-v]|version\s*)*` +
			`(\d+(?:\.\d+)*)`,
	)

	matches := re.FindStringSubmatch(lowerContent)
	if len(matches) > 1 {
		version := matches[1]
		return productKeyword + version
	}

	return ""
}

// 向结果集中添加指纹，优先保留更具体的（带版本）的指纹。
func updateCMS(matchedCMS map[string]struct{}, newFinger string) {
	var toDelete []string
	shouldAdd := true

	for existingFinger := range matchedCMS {
		lowerNew := strings.ToLower(newFinger)
		lowerExisting := strings.ToLower(existingFinger)

		if strings.HasPrefix(lowerNew, lowerExisting) && len(lowerNew) > len(lowerExisting) {
			toDelete = append(toDelete, existingFinger)
		} else if strings.HasPrefix(lowerExisting, lowerNew) && len(lowerExisting) > len(lowerNew) {
			shouldAdd = false
			break
		} else if lowerNew == lowerExisting && newFinger != existingFinger {
			toDelete = append(toDelete, existingFinger)
		}
	}

	if !shouldAdd {
		return
	}

	for _, key := range toDelete {
		delete(matchedCMS, key)
	}

	if _, exists := matchedCMS[newFinger]; !exists {
		matchedCMS[newFinger] = struct{}{}
	}
}
