package output

import (
	"encoding/csv"
	"os"
	"strconv"
)

// 结果保存到 CSV 文件
func SaveCSV(results []*Result, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"ID", "URL", "Code", "Length", "Title", "CMS"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// 写入数据
	for _, res := range results {
		record := []string{
			strconv.Itoa(res.ID),
			res.URL,
			strconv.Itoa(res.StatusCode),
			strconv.Itoa(res.ContentLength),
			res.Title,
			res.CMS,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}