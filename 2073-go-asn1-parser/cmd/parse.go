package cmd

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/asn1-parser/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	hexInput   string
	fileInput  string
	outputJSON bool
	isCert     bool
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "解析ASN.1数据或X.509证书",
	Long:  "解析DER编码的ASN.1数据或X.509证书，支持十六进制字符串或二进制文件输入",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runParse()
	},
}

func init() {
	parseCmd.Flags().StringVarP(&hexInput, "hex", "x", "", "十六进制字符串输入")
	parseCmd.Flags().StringVarP(&fileInput, "file", "f", "", "二进制文件输入")
	parseCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "输出JSON格式")
	parseCmd.Flags().BoolVarP(&isCert, "cert", "c", false, "解析为X.509证书")
}

func runParse() error {
	var data []byte
	var err error

	if hexInput == "" && fileInput == "" {
		return fmt.Errorf("请使用 --hex 或 --file 指定输入")
	}

	if hexInput != "" {
		data, err = parseHexString(hexInput)
		if err != nil {
			return err
		}
	} else {
		data, err = os.ReadFile(fileInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 文件不存在或无法读取: %s\n", fileInput)
			os.Exit(1)
		}
	}

	if isCert {
		return parseCertificate(data, outputJSON)
	}

	return parseASN1Data(data, outputJSON)
}

func parseHexString(hexStr string) ([]byte, error) {
	cleaned := strings.ReplaceAll(strings.ReplaceAll(hexStr, " ", ""), "\n", "")

	if len(cleaned)%2 != 0 {
		pos := len(cleaned)
		fmt.Fprintf(os.Stderr, "错误: 十六进制字符串长度为奇数，在位置 %d 处\n", pos)
		os.Exit(1)
	}

	data, err := hex.DecodeString(cleaned)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 十六进制字符串格式无效: %v\n", err)
		os.Exit(1)
	}

	return data, nil
}

func parseASN1Data(data []byte, outputJSON bool) error {
	options := parser.ParseOptions{
		MaxDepth: 100,
	}

	result, err := parser.ParseASN1(data, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析错误: %v\n", err)
		if result != nil {
			fmt.Fprintf(os.Stderr, "\n已解析部分:\n")
			if outputJSON {
				jsonOutput, jsonErr := result.ToJSON()
				if jsonErr == nil {
					fmt.Println(jsonOutput)
				}
			} else {
				fmt.Println(result.ToTreeString())
			}
		}
		os.Exit(1)
	}

	if outputJSON {
		jsonOutput, err := result.ToJSON()
		if err != nil {
			return fmt.Errorf("JSON编码错误: %v", err)
		}
		fmt.Println(jsonOutput)
	} else {
		fmt.Println(result.ToTreeString())
	}

	return nil
}

func parseCertificate(data []byte, outputJSON bool) error {
	cert, err := parser.ParseCertificate(data)
	if err != nil {
		return fmt.Errorf("证书解析错误: %v", err)
	}

	if outputJSON {
		jsonOutput, err := cert.ToJSON()
		if err != nil {
			return fmt.Errorf("JSON编码错误: %v", err)
		}
		fmt.Println(jsonOutput)
	} else {
		fmt.Println(cert.ToString())
	}

	return nil
}
