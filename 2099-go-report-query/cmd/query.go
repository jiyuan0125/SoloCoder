package cmd

import (
	"fmt"
	"os"
	"strings"

	"report-query/internal/db"
	"report-query/internal/output"
	"report-query/internal/query"
	"report-query/internal/storage"

	"github.com/spf13/cobra"
)

var (
	dbPath     string
	tableName  string
	fields     string
	conditions string
	whereClause string
	groupBy    string
	outFormat  string
	outFile    string
	saveAs     string
	useShortcut string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "执行查询",
	Long:  `执行数据库查询，支持通过条件表达式或自定义 WHERE 子句`,
	Run:   runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().StringVarP(&dbPath, "db", "d", "", "SQLite 数据库文件路径 (必需)")
	queryCmd.Flags().StringVarP(&tableName, "table", "t", "", "查询的表名")
	queryCmd.Flags().StringVarP(&fields, "fields", "f", "*", "要查询的字段，多个用逗号分隔")
	queryCmd.Flags().StringVarP(&conditions, "conditions", "c", "", "筛选条件，如: age>18 AND city='北京'")
	queryCmd.Flags().StringVarP(&whereClause, "where", "w", "", "自定义 WHERE 子句 (不包含 WHERE 关键字)")
	queryCmd.Flags().StringVarP(&groupBy, "group", "g", "", "GROUP BY 字段，多个用逗号分隔")
	queryCmd.Flags().StringVarP(&outFormat, "format", "F", "table", "输出格式: table 或 csv")
	queryCmd.Flags().StringVarP(&outFile, "output", "o", "", "CSV 输出文件路径")
	queryCmd.Flags().StringVarP(&saveAs, "save", "s", "", "将当前查询保存为快捷方式")
	queryCmd.Flags().StringVarP(&useShortcut, "use", "u", "", "使用已保存的快捷方式")
}

func runQuery(cmd *cobra.Command, args []string) {
	store, err := storage.NewStorage()
	if err != nil {
		exitWithError("初始化存储失败: "+err.Error(), 1)
	}

	if useShortcut != "" {
		sc, err := store.GetShortcut(useShortcut)
		if err != nil {
			exitWithError(err.Error(), 1)
		}
		fmt.Fprintf(os.Stderr, "使用快捷方式: %s\n", sc.Name)
		if dbPath == "" {
			dbPath = sc.DBPath
		}
		if tableName == "" {
			tableName = sc.Table
		}
		if cmd.Flags().Changed("fields") == false {
			fields = sc.Fields
		}
		if conditions == "" {
			conditions = sc.Conditions
		}
		if whereClause == "" {
			whereClause = sc.Where
		}
		if groupBy == "" {
			groupBy = sc.GroupBy
		}
	}

	if dbPath == "" {
		exitWithError("必须指定数据库路径 (--db 或 -d)", 1)
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		exitWithError("数据库文件不存在: "+dbPath, 1)
	}

	database, err := db.NewDB(dbPath)
	if err != nil {
		exitWithError(err.Error(), 1)
	}
	defer database.Close()

	if tableName == "" {
		tables, err := database.ListTables()
		if err != nil {
			exitWithError("获取表列表失败: "+err.Error(), 1)
		}
		fmt.Println("数据库中的表:")
		for _, t := range tables {
			fmt.Println("  -", t)
		}
		return
	}

	if !database.HasTable(tableName) {
		exitWithError("表不存在: "+tableName, 1)
	}

	qb := &query.QueryBuilder{
		DB:         database,
		Table:      tableName,
		Fields:     query.ParseFields(fields),
		Conditions: conditions,
		Where:      whereClause,
		GroupBy:    groupBy,
	}

	sqlQuery, _, err := qb.BuildAndValidate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:")
		exitWithError(err.Error(), 1)
	}

	fmt.Fprintf(os.Stderr, "执行 SQL: %s\n", sqlQuery)

	rows, err := database.ExecuteQuery(sqlQuery)
	if err != nil {
		exitWithError("查询执行失败: "+err.Error(), 1)
	}
	defer rows.Close()

	result, err := output.ReadRows(rows)
	if err != nil {
		exitWithError("读取结果失败: "+err.Error(), 1)
	}

	if len(result.Rows) == 0 {
		fmt.Println("无匹配数据")
	} else {
		if err := output.Output(result, outFormat, outFile); err != nil {
			exitWithError(err.Error(), 1)
		}
	}

	if err := store.AddHistory(sqlQuery); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 无法保存查询历史: %v\n", err)
	}

	if saveAs != "" {
		if useShortcut != "" && useShortcut == saveAs {
			fmt.Fprintf(os.Stderr, "警告: 快捷方式名不能与正在使用的快捷方式相同\n")
		} else {
			sc := storage.Shortcut{
				Name:       saveAs,
				DBPath:     dbPath,
				Table:      tableName,
				Fields:     strings.Join(qb.Fields, ", "),
				Conditions: conditions,
				Where:      whereClause,
				GroupBy:    groupBy,
			}
			if err := store.AddShortcut(sc); err != nil {
				fmt.Fprintf(os.Stderr, "警告: 无法保存快捷方式: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "已保存快捷方式: %s\n", saveAs)
			}
		}
	}
}
