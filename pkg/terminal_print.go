package pkg

import (
	"GoWebTrace/internal/output"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
)

func TerminalPrint(results []*output.Result) {
	log.Println(color.CyanString("\n\n[*] GoWebTrace 指纹识别结果摘要:"))

	const urlWidth, statusWidth, titleWidth, cmsWidth = 40, 8, 30, 40

	// 为字符串添加空格，直到达到指定宽度
	pad := func(s string, width int) string {
		sWidth := runewidth.StringWidth(s)
		if sWidth >= width {
			return s
		}
		return s + strings.Repeat(" ", width-sWidth)
	}

	// 打印表头
	header := fmt.Sprintf("%s %s %s %s",
		pad("URL", urlWidth),
		pad("Status", statusWidth),
		pad("Title", titleWidth),
		pad("CMS", cmsWidth))
	fmt.Println(color.CyanString(header))

	separator := fmt.Sprintf("%s %s %s %s",
		strings.Repeat("-", urlWidth),
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", titleWidth),
		strings.Repeat("-", cmsWidth))
	fmt.Println(color.CyanString(separator))

	for _, res := range results {
		// 状态码颜色
		var statusColor *color.Color
		hasCMS := res.CMS != "None" && res.CMS != ""
		switch {
		case res.StatusCode >= 300:
			statusColor = color.New(color.FgRed)
		case res.StatusCode >= 200:
			statusColor = color.New(color.FgGreen)
		default:
			statusColor = color.New(color.FgWhite)
		}

		// CMS显示内容和颜色
		cmsDisplay := "N/A"
		if hasCMS {
			cmsDisplay = res.CMS
		}

		urlRunes := []rune(res.URL)
		titleRunes := []rune(res.Title)
		cmsRunes := []rune(cmsDisplay)

		for i := 0; ; i++ {
			urlChunk := getChunk(&urlRunes, urlWidth)
			titleChunk := getChunk(&titleRunes, titleWidth)
			cmsChunk := getChunk(&cmsRunes, cmsWidth)

			if urlChunk == "" && titleChunk == "" && cmsChunk == "" {
				break
			}

			var statusStr string
			if i == 0 {
				statusStr = fmt.Sprintf("%d", res.StatusCode)
			}

			// 构造无颜色行
			line := fmt.Sprintf("%s %s %s %s",
				pad(urlChunk, urlWidth),
				pad(statusStr, statusWidth),
				pad(titleChunk, titleWidth),
				pad(cmsChunk, cmsWidth))

			fmt.Println(statusColor.Sprint(line))
		}
	}
	fmt.Println()
}

// 修改原始切片，移除已提取的块
func getChunk(runes *[]rune, width int) string {
	if len(*runes) == 0 {
		return ""
	}

	currentWidth := 0
	end := 0
	for i, r := range *runes {
		charWidth := runewidth.RuneWidth(r)
		if currentWidth+charWidth > width {
			break
		}
		currentWidth += charWidth
		end = i + 1
	}

	// 中文强制截取一个字符
	if end == 0 && len(*runes) > 0 {
		end = 1
	}

	chunk := string((*runes)[:end])
	*runes = (*runes)[end:]
	return chunk
}
