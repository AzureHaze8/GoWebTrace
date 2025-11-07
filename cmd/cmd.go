package cmd

import (
	"GoWebTrace/internal/engine"
	"GoWebTrace/internal/network"
	"GoWebTrace/internal/output"
	"GoWebTrace/pkg"
	"flag"
	"fmt"
	"log"

	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"time"

	"github.com/fatih/color"
)

// 处理单个URL的扫描逻辑
func processURL(targetURL string, rules *network.LocationRules, id int, showCert bool, proxy string) *output.Result {
	targetURL = pkg.Trim(targetURL)
	info, err := network.SendRequest(targetURL, showCert, proxy)
	if err != nil {
		log.Printf(color.RedString("[!]请求失败: %v\n"), err)
		return nil
	}

	// log.Println("正在解析HTML...")
	info.ParseHTML(proxy)

	if info.Favicon != "" {
		// log.Println("正在获取Favicon哈希...")
		faviconMmh3, faviconMd5, err := network.GetFaviconHash(info.Favicon, proxy)
		if err != nil {
			log.Printf(color.RedString("[!]获取Favicon哈希失败: %v\n"), err)
		}
		info.FaviconMmh3 = faviconMmh3
		info.FaviconMd5 = faviconMd5
	}

	// 指纹识别和结果打印
	// log.Println("正在进行指纹识别...")
	matchedCMS := engine.MatchFingerprints(info, rules)
	var cmsList []string
	if len(matchedCMS) > 0 {
		for cms := range matchedCMS {
			cmsList = append(cmsList, cms)
		}
	}
	// 识别结果
	if len(cmsList) > 0 {
		log.Printf(color.GreenString("[+]正在识别:%s ,已匹配指纹: %s\n"), info.URL, strings.Join(cmsList, ", "))
	} else {
		log.Printf(color.YellowString("[-]正在识别:%s ,未匹配到指纹。\n"), info.URL)
	}

	screenshotPath, err := output.SaveScreenshot(info.URL)
	if err != nil {
		log.Printf(color.RedString("[!]%s 截图失败，错误信息: %v"), info.URL, err)
		screenshotPath = ""
	}

	cmsString := "None"
	if len(cmsList) > 0 {
		cmsString = strings.Join(cmsList, ", ")
	}

	return output.NewResult(
		id,
		info.URL,
		info.StatusCode,
		len(info.Body),
		info.Title,
		cmsString,
		screenshotPath,
	)
}

func Execute() {
	options := pkg.Options{}
	flag.StringVar(&options.TargetURL, "u", "", "URL,单个URL")
	flag.StringVar(&options.FilePath, "f", "", "filePath,URL列表的文件路径")
	flag.IntVar(&options.Concurrency, "c", 5, "concurrency,并发数")
	flag.StringVar(&options.RulePath, "r", "config/finger.json", "rulePath,指纹规则文件")
	flag.StringVar(&options.Output, "o", "", "output,输出文件,只支持.csv和.html格式(如:output.csv)\n多个文件用逗号分隔(如:output.csv,output.html)")
	flag.BoolVar(&options.CertFinger, "cert", false, "certFinger,启用证书指纹识别,默认false")
	flag.BoolVar(&options.ProxyList, "pl", false, "proxyList,是否使用代理列表，默认false")
	flag.StringVar(&options.ProxyURL, "p", "", "proxyURL,单个代理URL，支持http/https/socks4/socks5\n(如:http://127.0.0.1:1080或socks5://127.0.0.1:1080)")
	flag.Parse()

	// 检查是否提供了目标URL或文件路径
	if (options.TargetURL == "" && options.FilePath == "") || (options.TargetURL != "" && options.FilePath != "") {
		log.Fatal(color.RedString("[!]请输入(-h)查看帮助\n"))
	}
	if options.ProxyList && options.ProxyURL != "" {
		log.Fatal(color.RedString("[!]不能同时使用代理列表 (-pl) 和单个代理 (-p)。\n"))
	}

	// 初始化代理管理器
	proxyManager, err := pkg.New(options.ProxyList, options.ProxyURL)
	if err != nil {
		log.Fatalf(color.RedString("[!]初始化代理管理器失败: %v\n"), err)
	}

	// 加载规则文件
	rules, err := engine.RuleAnalyzer(options.RulePath)
	if err != nil {
		log.Fatalf(color.RedString("[!]加载规则文件失败: %v\n"), err)
	}

	// 在扫描开始时记录时间
	startTime := time.Now()

	// 获取URL列表
	var urls []string
	if options.TargetURL != "" {
		urls = []string{options.TargetURL}
	} else {
		urls, err = pkg.ReadLines(options.FilePath)
		if err != nil {
			log.Fatalf(color.RedString("[!]从文件 %s 读取URL失败: %v\n"), options.FilePath, err)
		}
	}

	if len(urls) == 0 {
		log.Fatal(color.YellowString("[-]没有要处理的URL，请输入目标URL (-u) 或包含URL列表的文件路径 (-f)\n"))
	}

	// 并发处理URL
	var wg sync.WaitGroup
	resultsChan := make(chan *output.Result, len(urls))
	semaphore := make(chan struct{}, options.Concurrency)

	for i, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(u string, id int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			// 从管理器获取代理
			currentProxy := proxyManager.GetProxy()
			if result := processURL(u, rules, id, options.CertFinger, currentProxy); result != nil {
				resultsChan <- result
			}
		}(url, i+1)
	}

	wg.Wait()
	close(resultsChan)

	var results []*output.Result
	for result := range resultsChan {
		results = append(results, result)
	}
	// 按ID对结果进行排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})

	pkg.TerminalPrint(results)

	for i := range results {
		results[i].ID = i + 1
	}
	// 文件输出
	if options.Output != "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf(color.RedString("[!]无法获取当前工作目录: %v\n"), err)
		}

		outputFiles := strings.Split(options.Output, ",")

		for _, file := range outputFiles {
			outputFilePath := filepath.Join(wd, strings.TrimSpace(file))
			if strings.HasSuffix(outputFilePath, ".csv") {
				output.SaveCSV(results, outputFilePath)
				fmt.Printf(color.GreenString("[*]扫描完成！CSV结果已保存至：  %s\n"), outputFilePath)
			} else if strings.HasSuffix(outputFilePath, ".html") {
				output.SaveHTML(results, outputFilePath)
				fmt.Printf(color.GreenString("[*]扫描完成！HTML结果已保存至： %s\n"), outputFilePath)
			}
		}
	} else {
		fmt.Println(color.YellowString("[-]未指定输出文件 (-o)，结果未保存。\n"))
	}

	// 计算并打印扫描耗时
	elapsed := time.Since(startTime)
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	fmt.Printf(color.GreenString("扫描耗时: %d分%d秒"), minutes, seconds)
}
