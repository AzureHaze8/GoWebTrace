package pkg

import (
	"bufio"
	"os"
	"strings"
)

// 移除字符串首尾的空白字符
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// 自动去除每行前后的空白字符，并忽略空行
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
