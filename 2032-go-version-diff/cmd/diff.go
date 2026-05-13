package cmd

import (
	"config-diff/internal/comparator"
	"config-diff/internal/git"
	"config-diff/internal/output"
	"config-diff/internal/parser"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repoPath    string
	oldRef      string
	newRef      string
	filePath    string
	outputFormat string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "对比两个版本之间的配置文件差异",
	Long:  `对比 Git 仓库中两个版本（分支或 commit）之间指定配置文件的差异。`,
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&repoPath, "repo", "r", "", "Git 仓库路径（也可通过 CONFIG_DIFF_REPO 环境变量指定）")
	diffCmd.Flags().StringVarP(&oldRef, "old", "o", "", "旧版本（分支名或 commit hash）")
	diffCmd.Flags().StringVarP(&newRef, "new", "n", "", "新版本（分支名或 commit hash）")
	diffCmd.Flags().StringVarP(&filePath, "file", "f", "", "配置文件在仓库中的相对路径")
	diffCmd.Flags().StringVarP(&outputFormat, "format", "F", "terminal", "输出格式：terminal 或 json")

	diffCmd.MarkFlagRequired("old")
	diffCmd.MarkFlagRequired("new")
	diffCmd.MarkFlagRequired("file")
}

func runDiff(cmd *cobra.Command, args []string) error {
	if repoPath == "" {
		repoPath = os.Getenv("CONFIG_DIFF_REPO")
	}

	if repoPath == "" {
		return fmt.Errorf("请通过 --repo 参数或 CONFIG_DIFF_REPO 环境变量指定仓库路径")
	}

	repo, err := git.NewRepository(repoPath)
	if err != nil {
		return err
	}

	oldContent, err := repo.GetFileContent(oldRef, filePath)
	if err != nil {
		return err
	}

	newContent, err := repo.GetFileContent(newRef, filePath)
	if err != nil {
		return err
	}

	oldData, err := parser.ParseFile(filePath, oldContent)
	if err != nil {
		return err
	}

	newData, err := parser.ParseFile(filePath, newContent)
	if err != nil {
		return err
	}

	changes := comparator.Compare(oldData, newData)

	switch outputFormat {
	case "json":
		jsonOutput, err := output.FormatJSON(changes)
		if err != nil {
			return err
		}
		fmt.Println(jsonOutput)
	case "terminal":
		fmt.Print(output.FormatTerminal(changes))
	default:
		return fmt.Errorf("不支持的输出格式: %s，支持的格式：terminal, json", outputFormat)
	}

	return nil
}
