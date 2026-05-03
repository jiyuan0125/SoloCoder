package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"sqlmigrate/sqlmigrate"
)

// 定义命令行参数
var (
	dir     = flag.String("dir", "./migrations", "迁移文件目录")
	command = flag.String("command", "up", "命令: up 或 down")
	target  = flag.String("target", "", "目标版本 (可选)")
	dryRun  = flag.Bool("dry-run", false, "只打印 SQL 不实际执行")
	dsn     = flag.String("dsn", "", "数据库连接字符串 (MySQL DSN format: user:password@tcp(localhost:3306)/dbname)")
)

func main() {
	flag.Parse()

	// 验证参数
	if *dsn == "" {
		fmt.Println("错误: 必须提供 -dsn 参数")
		os.Exit(1)
	}

	if *command != "up" && *command != "down" {
		fmt.Println("错误: 无效的命令，必须是 up 或 down")
		os.Exit(1)
	}

	// 连接数据库
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		fmt.Printf("数据库 ping 失败: %v\n", err)
		os.Exit(1)
	}

	// 获取文件锁 (使用 flock)
	lock, err := acquireFileLock(*dir)
	if err != nil {
		fmt.Printf("获取文件锁失败: %v\n", err)
		os.Exit(1)
	}
	defer releaseFileLock(lock)

	// 执行迁移
	if err := runMigrations(db, *dir, *command, *target, *dryRun); err != nil {
		fmt.Printf("迁移失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("迁移完成")
}

func runMigrations(db *sql.DB, dir, command, target string, dryRun bool) error {
	// 创建 scanner 并扫描迁移文件
	scanner := sqlmigrate.NewScanner(dir)
	migrations, err := scanner.Scan()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		fmt.Println("没有找到迁移文件")
		return nil
	}

	// 创建 executor
	executor := sqlmigrate.NewExecutor(db, dryRun)

	// 确保 schema_migrations 表存在
	if err := executor.EnsureSchemaMigrationsTable(); err != nil {
		return err
	}

	// 获取当前已应用的版本
	appliedVersions, err := executor.GetAppliedVersions()
	if err != nil {
		return err
	}

	// 根据命令执行迁移
	switch command {
	case "up":
		return runUp(migrations, executor, appliedVersions, target)
	case "down":
		return runDown(migrations, executor, appliedVersions, target)
	}

	return nil
}

func runUp(migrations []*sqlmigrate.Migration, executor *sqlmigrate.Executor, appliedVersions map[string]bool, targetVersion string) error {
	for _, migration := range migrations {
		// 如果指定了目标版本，只执行到目标版本
		if targetVersion != "" && migration.Version > targetVersion {
			break
		}

		// 跳过已应用的版本
		if appliedVersions[migration.Version] {
			fmt.Printf("跳过已应用的迁移: %s\n", migration.Version)
			continue
		}

		// 执行迁移
		if err := executor.ExecuteUp(migration); err != nil {
			return err
		}
	}
	return nil
}

func runDown(migrations []*sqlmigrate.Migration, executor *sqlmigrate.Executor, appliedVersions map[string]bool, targetVersion string) error {
	// 如果指定了目标版本
	if targetVersion != "" {
		// 检查目标版本是否存在
		targetExists := false
		for _, migration := range migrations {
			if migration.Version == targetVersion {
				targetExists = true
				break
			}
		}

		if !targetExists {
			return fmt.Errorf("目标版本 %s 不存在", targetVersion)
		}

		// 倒序回滚到目标版本
		for i := len(migrations) - 1; i >= 0; i-- {
			migration := migrations[i]

			// 如果已经到达目标版本，停止
			if migration.Version == targetVersion {
				break
			}

			// 只回滚已应用的版本
			if appliedVersions[migration.Version] {
				if err := executor.ExecuteDown(migration); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// 没有指定目标版本，只回滚最后一个已应用的版本
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if appliedVersions[migration.Version] {
			if err := executor.ExecuteDown(migration); err != nil {
				return err
			}
			break // 只回滚一个版本
		}
	}
	return nil
}

// 文件锁相关函数 - 使用 flock
type FileLock struct {
	file *os.File
	path string
}

func acquireFileLock(dir string) (*FileLock, error) {
	lockPath := dir + "/.migrate.lock"

	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 打开锁文件 (如果不存在则创建)
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("无法打开锁文件: %w", err)
	}

	// 尝试获取排他锁，最多等待 5 秒
	timeout := 5 * time.Second
	start := time.Now()

	for {
		// 尝试获取排他锁 (非阻塞)
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			// 成功获取锁
			return &FileLock{file: file, path: lockPath}, nil
		}

		// 检查是否是因为锁被占用
		if err != syscall.EWOULDBLOCK {
			file.Close()
			return nil, fmt.Errorf("获取锁失败: %w", err)
		}

		// 检查是否超时
		if time.Since(start) >= timeout {
			file.Close()
			return nil, fmt.Errorf("获取文件锁超时 (等待 %v)", timeout)
		}

		// 等待 100ms 后重试
		time.Sleep(100 * time.Millisecond)
	}
}

func releaseFileLock(lock *FileLock) error {
	if lock == nil {
		return nil
	}

	// 释放锁
	if lock.file != nil {
		// 释放 flock 锁
		syscall.Flock(int(lock.file.Fd()), syscall.LOCK_UN)
		// 关闭文件
		lock.file.Close()
	}

	return nil
}
