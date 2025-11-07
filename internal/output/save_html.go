package output

import (
	"html/template"
	"os"
)

// 结果保存到 HTML 文件
func SaveHTML(results []*Result, filename string) error {
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, results)
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <title>GoWebTrace Scan Report</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif;
            background-color: #f0f2f5;
            color: #333;
            margin: 0;
            padding: 20px;
            font-size: 14px;
        }
        .container {
            max-width: 1400px;
            margin: 20px auto;
            background-color: #fff;
            border-radius: 10px;
            box-shadow: 0 6px 18px rgba(0,0,0,0.07);
            overflow: hidden;
        }
        header {
            background: linear-gradient(to right, #2c3e50, #34495e);
            color: #ecf0f1;
            padding: 25px 30px;
            text-align: center;
        }
        header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 600;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            text-align: left;
        }
        th, td {
            padding: 15px 20px;
            border-bottom: 1px solid #e6e9ed;
        }
        th {
            background-color: #f5f7fa;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 12px;
            color: #555;
        }
        tr:last-child td {
            border-bottom: none;
        }
        tr:hover {
            background-color: #f9fafb;
        }
        a {
            color: #3498db;
            text-decoration: none;
            transition: color 0.2s;
        }
        .url-cell {
            max-width: 300px; 
            word-break: break-all;
        }
        a:hover {
            color: #2980b9;
            text-decoration: underline;
        }
        .screenshot-thumb {
            max-width: 200px;
            height: auto;
            border-radius: 6px;
            cursor: pointer;
            transition: transform 0.2s;
        }
        .screenshot-thumb:hover {
            transform: scale(1.05);
        }
        .screenshot-cell {
            text-align: center;
            vertical-align: middle;
        }
        .status-code {
            font-weight: bold;
        }
        footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #999;
            background-color: #f5f7fa;
        }

        /* Lightbox (Modal) styles */
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            overflow: auto;
            background-color: rgba(0,0,0,0.85);
            justify-content: center;
            align-items: center;
        }
        .modal-content {
            margin: auto;
            display: block;
            max-width: 90%;
            max-height: 90%;
        }
        .modal-close {
            position: absolute;
            top: 20px;
            right: 35px;
            color: #f1f1f1;
            font-size: 40px;
            font-weight: bold;
            transition: 0.3s;
            cursor: pointer;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>GoWebTrace Scan Report</h1>
        </header>
        <table>
            <thead>
                <tr>
                    <th>ID</th>
                    <th>URL</th>
                    <th>Status</th>
                    <th>Length</th>
                    <th>Title</th>
                    <th>CMS</th>
                    <th>Timestamp</th>
                    <th class="screenshot-cell">Screenshot</th>
                </tr>
            </thead>
            <tbody>
                {{range .}}
                <tr>
                    <td>{{.ID}}</td>
                    <td class="url-cell"><a href="{{.URL}}" target="_blank">{{.URL}}</a></td>
                    <td class="status-code">{{.StatusCode}}</td>
                    <td>{{.ContentLength}} bytes</td>
                    <td>{{.Title}}</td>
                    <td>{{.CMS}}</td>
                    <td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
                    <td class="screenshot-cell">
                        {{if .ScreenshotPath}}
                        <img src="{{.ScreenshotPath}}" alt="Screenshot for {{.URL}}" class="screenshot-thumb" onclick="showModal('{{.ScreenshotPath}}')">
                        {{else}}
                        N/A
                        {{end}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
        <footer>
            Report generated by GoWebTrace
        </footer>
    </div>

    <!-- The Modal -->
    <div id="myModal" class="modal" onclick="this.style.display='none'">
        <span class="modal-close">&times;</span>
        <img class="modal-content" id="img01">
    </div>

    <script>
        var modal = document.getElementById('myModal');
        var modalImg = document.getElementById("img01");
        function showModal(src) {
            modal.style.display = "flex";
            modalImg.src = src;
        }
        document.querySelector('.modal-close').onclick = function() {
            modal.style.display = "none";
        }
    </script>
</body>
</html>
`
