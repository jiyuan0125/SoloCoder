package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	StopGracePeriod = 5 * time.Second
)

func stopService(binaryPath string) error {
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}

	processName := filepath.Base(absPath)
	pids, err := findProcesses(processName)
	if err != nil {
		return fmt.Errorf("查找进程失败: %v", err)
	}

	if len(pids) == 0 {
		Debug("未找到运行中的进程: %s", processName)
		return nil
	}

	Info("找到 %d 个正在运行的进程: %v", len(pids), pids)

	for _, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			Debug("找不到进程 %d: %v", pid, err)
			continue
		}

		Info("发送 SIGTERM 到进程 %d", pid)
		if err := proc.Signal(os.Interrupt); err != nil {
			if runtime.GOOS == "windows" {
				Debug("Windows系统不支持SIGTERM，尝试直接kill: %v", err)
			} else {
				Debug("发送 SIGTERM 失败: %v", err)
			}
		}
	}

	Info("等待 %v 让进程优雅退出", StopGracePeriod)
	deadline := time.Now().Add(StopGracePeriod)

	for time.Now().Before(deadline) {
		remaining := deadline.Sub(time.Now())
		Debug("检查进程是否已退出，剩余等待时间: %v", remaining)

		allGone := true
		for _, pid := range pids {
			if isProcessRunning(pid) {
				allGone = false
				break
			}
		}

		if allGone {
			Info("所有进程已优雅退出")
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	Info("优雅退出超时，发送 SIGKILL")
	for _, pid := range pids {
		if isProcessRunning(pid) {
			proc, err := os.FindProcess(pid)
			if err != nil {
				Debug("找不到进程 %d: %v", pid, err)
				continue
			}
			Info("发送 SIGKILL 到进程 %d", pid)
			if err := proc.Kill(); err != nil {
				Debug("发送 SIGKILL 失败: %v", err)
			}
		}
	}

	time.Sleep(1 * time.Second)

	for _, pid := range pids {
		if isProcessRunning(pid) {
			return fmt.Errorf("进程 %d 仍然在运行，无法停止", pid)
		}
	}

	Info("所有进程已停止")
	return nil
}

func startService(binaryPath string) error {
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("二进制文件不存在: %s", absPath)
	}

	cmd := exec.Command(absPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = getSysProcAttr()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动服务失败: %v", err)
	}

	Info("服务已启动，PID: %d", cmd.Process.Pid)

	time.Sleep(1 * time.Second)

	if !isProcessRunning(cmd.Process.Pid) {
		return fmt.Errorf("服务启动后立即退出，PID: %d", cmd.Process.Pid)
	}

	return nil
}

func findProcesses(processName string) ([]int, error) {
	if runtime.GOOS == "windows" {
		return findProcessesWindows(processName)
	}

	return findProcessesUnix(processName)
}

func findProcessesUnix(processName string) ([]int, error) {
	var pids []int

	cmd := exec.Command("pgrep", "-f", processName)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return pids, nil
			}
		}
		return nil, fmt.Errorf("执行 pgrep 失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if pid, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && pid > 0 {
			pids = append(pids, pid)
		}
	}

	return pids, nil
}

func findProcessesWindows(processName string) ([]int, error) {
	var pids []int

	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName), "/NH")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 tasklist 失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if pid, err := strconv.Atoi(parts[1]); err == nil && pid > 0 {
				pids = append(pids, pid)
			}
		}
	}

	return pids, nil
}

func isProcessRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		return isProcessRunningWindows(pid)
	}

	if err := proc.Signal(nil); err != nil {
		return false
	}
	return true
}

func isProcessRunningWindows(pid int) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.Contains(line, strconv.Itoa(pid)) {
			return true
		}
	}

	return false
}
