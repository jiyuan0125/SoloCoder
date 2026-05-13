package output

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"reportgen/internal/model"
)

func GenerateHTMLReport(result *model.ReportResult, outputPath string) error {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>`)
	buf.WriteString(escapeHTML(result.Title))
	buf.WriteString(`</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            margin: 40px;
            background-color: #f5f5f5;
        }
        .report-container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #4a90d9;
            padding-bottom: 15px;
            margin-bottom: 30px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 20px;
        }
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid #e0e0e0;
        }
        th {
            background-color: #4a90d9;
            color: white;
            font-weight: 600;
            text-transform: uppercase;
            font-size: 13px;
            letter-spacing: 0.5px;
        }
        tr:hover {
            background-color: #f9f9f9;
        }
        tr:nth-child(even) {
            background-color: #fafafa;
        }
        .anomaly {
            color: #e74c3c;
            font-weight: bold;
            background-color: #ffebee;
        }
        .anomaly::before, .anomaly::after {
            content: '*';
        }
        .summary {
            margin-top: 20px;
            padding: 15px;
            background-color: #e8f5e9;
            border-radius: 4px;
            color: #2e7d32;
        }
        .metadata {
            color: #666;
            font-size: 14px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="report-container">
`)

	if result.Title != "" {
		buf.WriteString(fmt.Sprintf("        <h1>%s</h1>\n", escapeHTML(result.Title)))
	}

	buf.WriteString(fmt.Sprintf(`        <div class="metadata">生成时间: %s</div>`+"\n", time.Now().Format("2006-01-02 15:04:05")))

	if len(result.Groups) > 0 {
		renderGroupedHTML(&buf, result)
	} else {
		renderSimpleHTML(&buf, result)
	}

	if result.HasAnomaly {
		buf.WriteString(`        <div class="summary">报告包含异常值，已用红色背景和星号标记。</div>` + "\n")
	}

	buf.WriteString(`    </div>
</body>
</html>
`)

	if outputPath == "" {
		fmt.Println(buf.String())
		return nil
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func renderSimpleHTML(buf *bytes.Buffer, result *model.ReportResult) {
	if len(result.Headers) == 0 {
		return
	}

	buf.WriteString("        <table>\n")
	buf.WriteString("            <thead>\n                <tr>\n")
	for _, h := range result.Headers {
		buf.WriteString(fmt.Sprintf("                    <th>%s</th>\n", escapeHTML(h)))
	}
	buf.WriteString("                </tr>\n            </thead>\n")
	buf.WriteString("            <tbody>\n")

	for _, row := range result.Rows {
		buf.WriteString("                <tr>\n")
		for _, h := range result.Headers {
			pv, exists := row.Values[h]
			if !exists {
				buf.WriteString("                    <td></td>\n")
				continue
			}
			buf.WriteString(fmt.Sprintf("                    <td%s>%s</td>\n",
				getAnomalyClass(pv.IsAnomaly),
				escapeHTML(formatValueSimple(pv.Value))))
		}
		buf.WriteString("                </tr>\n")
	}

	buf.WriteString("            </tbody>\n        </table>\n")
}

func renderGroupedHTML(buf *bytes.Buffer, result *model.ReportResult) {
	if len(result.Groups) == 0 {
		return
	}

	groupKeys := make([]string, 0)
	if len(result.Groups) > 0 {
		for k := range result.Groups[0].Key.Fields {
			groupKeys = append(groupKeys, k)
		}
	}

	aggKeys := make([]string, 0)
	if len(result.Groups) > 0 {
		for k := range result.Groups[0].Values {
			aggKeys = append(aggKeys, k)
		}
	}

	headers := make([]string, 0, len(groupKeys)+len(aggKeys))
	headers = append(headers, groupKeys...)
	headers = append(headers, aggKeys...)

	buf.WriteString("        <table>\n")
	buf.WriteString("            <thead>\n                <tr>\n")
	for _, h := range headers {
		buf.WriteString(fmt.Sprintf("                    <th>%s</th>\n", escapeHTML(h)))
	}
	buf.WriteString("                </tr>\n            </thead>\n")
	buf.WriteString("            <tbody>\n")

	for _, group := range result.Groups {
		buf.WriteString("                <tr>\n")
		for _, k := range groupKeys {
			buf.WriteString(fmt.Sprintf("                    <td>%s</td>\n",
				escapeHTML(fmt.Sprintf("%v", group.Key.Fields[k]))))
		}
		for _, k := range aggKeys {
			buf.WriteString(fmt.Sprintf("                    <td>%s</td>\n",
				escapeHTML(formatValueSimple(group.Values[k]))))
		}
		buf.WriteString("                </tr>\n")
	}

	buf.WriteString("            </tbody>\n        </table>\n")
}

func getAnomalyClass(isAnomaly bool) string {
	if isAnomaly {
		return ` class="anomaly"`
	}
	return ""
}

func escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}
