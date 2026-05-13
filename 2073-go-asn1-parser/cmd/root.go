package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "asn1-parser",
	Short: "ASN.1解析系统管理后台",
	Long:  "ASN.1解析系统管理后台，支持解析DER编码的ASN.1数据和X.509证书",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
