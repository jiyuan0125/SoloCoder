package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const PIDFile = "gocron.pid"

func IsDaemonMode() bool {
	return os.Getenv("GOCRON_DAEMON") == "1"
}

func StartDaemon() error {
	if IsDaemonMode() {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	args := os.Args[1:]
	daemonArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg != "-d" && arg != "--daemon" {
			daemonArgs = append(daemonArgs, arg)
		}
	}

	cmd := exec.Command(exe, daemonArgs...)
	cmd.Env = append(os.Environ(), "GOCRON_DAEMON=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return err
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(PIDFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return err
	}

	fmt.Printf("gocron 守护进程已启动，PID: %d\n", pid)
	fmt.Printf("PID 文件: %s\n", filepath.Join(".", PIDFile))
	os.Exit(0)
	return nil
}

func StopDaemon() error {
	data, err := os.ReadFile(PIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("未找到 PID 文件，守护进程可能未启动")
		}
		return err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("PID 文件格式错误: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("无法终止进程 (PID: %d): %v", pid, err)
	}

	if err := os.Remove(PIDFile); err != nil {
		fmt.Printf("警告: 无法删除 PID 文件: %v\n", err)
	}

	fmt.Printf("gocron 守护进程已停止 (PID: %d)\n", pid)
	return nil
}

func Status() error {
	data, err := os.ReadFile(PIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("gocron 守护进程未运行")
			return nil
		}
		return err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("PID 文件格式错误: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("gocron 守护进程 PID: %d (进程不存在)\n", pid)
		return nil
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		fmt.Printf("gocron 守护进程 PID: %d (进程已停止)\n", pid)
		return nil
	}

	fmt.Printf("gocron 守护进程正在运行，PID: %d\n", pid)
	return nil
}
