package lock

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func NewLock(targetFilePath string) *Lock {
	return &Lock{
		filePath:     targetFilePath,
		lockFilePath: targetFilePath + LockFileSuffix,
		openFiles:    make([]*os.File, 0),
	}
}

func (l *Lock) LockExclusive(timeout time.Duration) error {
	return l.acquire(LockModeExclusive, timeout)
}

func (l *Lock) LockShared(timeout time.Duration) error {
	return l.acquire(LockModeShared, timeout)
}

func (l *Lock) Unlock() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.holdCount <= 0 {
		return ErrLockNotHeld
	}

	l.holdCount--

	if l.holdCount > 0 {
		return nil
	}

	if len(l.openFiles) == 0 {
		l.heldMode = nil
		return nil
	}

	firstFile := l.openFiles[0]

	if err := writeHolderInfo(firstFile, nil); err != nil {
		for _, f := range l.openFiles {
			_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
			_ = f.Close()
		}
		l.openFiles = nil
		l.heldMode = nil
		l.holdCount = 0
		return err
	}

	for _, f := range l.openFiles {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}

	l.openFiles = nil
	l.heldMode = nil
	l.holdCount = 0
	return nil
}

func (l *Lock) acquire(mode LockMode, timeout time.Duration) error {
	if mode != LockModeExclusive && mode != LockModeShared {
		return ErrInvalidMode
	}

	config := DefaultConfig()

	var actualTimeout time.Duration
	noWait := false

	switch {
	case timeout == 0:
		noWait = true
		actualTimeout = 0
	case timeout < 0:
		actualTimeout = config.Timeout
	default:
		actualTimeout = timeout
	}

	l.mu.Lock()

	if l.holdCount > 0 && l.heldMode != nil {
		held := *l.heldMode
		compatible := false
		if held == LockModeExclusive || held == mode {
			compatible = true
		}

		if compatible {
			l.holdCount++
			l.mu.Unlock()
			return nil
		}

		l.mu.Unlock()
		return ErrInvalidMode
	}

	l.mu.Unlock()

	if noWait {
		return l.doAcquireNoWait(mode)
	}

	return l.doAcquireWithTimeout(mode, actualTimeout, config.MaxWaitDuration)
}

func (l *Lock) doAcquireNoWait(mode LockMode) error {
	f, err := os.OpenFile(l.lockFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	lockHow := syscall.LOCK_SH
	if mode == LockModeExclusive {
		lockHow = syscall.LOCK_EX
	}

	err = syscall.Flock(int(f.Fd()), lockHow|syscall.LOCK_NB)
	if err != nil {
		_ = f.Close()
		if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
			return ErrLockTimeout
		}
		return fmt.Errorf("flock failed: %w", err)
	}

	holderInfo, err := readAndCheckHolder(f)
	if err != nil {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
		return err
	}

	newHolder := &LockHolderInfo{
		PID:       os.Getpid(),
		StartTime: time.Now(),
		Mode:      modeToString(mode),
	}

	if err := writeHolderInfo(f, newHolder); err != nil {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
		return err
	}

	modeCopy := mode

	l.mu.Lock()
	l.heldMode = &modeCopy
	l.holdCount = 1
	l.openFiles = append(l.openFiles, f)
	l.mu.Unlock()

	_ = holderInfo
	return nil
}

func (l *Lock) doAcquireWithTimeout(mode LockMode, timeout, maxWait time.Duration) error {
	startTime := time.Now()

	lockHow := syscall.LOCK_SH
	if mode == LockModeExclusive {
		lockHow = syscall.LOCK_EX
	}

	f, err := os.OpenFile(l.lockFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	attemptStart := time.Now()
	for {
		elapsedTotal := time.Since(startTime)
		if elapsedTotal > maxWait {
			_ = f.Close()
			return ErrDeadlockDetected
		}

		elapsedAttempt := time.Since(attemptStart)
		if elapsedAttempt >= timeout {
			_ = f.Close()
			return ErrLockTimeout
		}

		err := syscall.Flock(int(f.Fd()), lockHow|syscall.LOCK_NB)
		if err == nil {
			holderInfo, err := readAndCheckHolder(f)
			if err != nil {
				_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				_ = f.Close()
				return err
			}

			newHolder := &LockHolderInfo{
				PID:       os.Getpid(),
				StartTime: time.Now(),
				Mode:      modeToString(mode),
			}

			if err := writeHolderInfo(f, newHolder); err != nil {
				_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				_ = f.Close()
				return err
			}

			modeCopy := mode

			l.mu.Lock()
			l.heldMode = &modeCopy
			l.holdCount = 1
			l.openFiles = append(l.openFiles, f)
			l.mu.Unlock()

			_ = holderInfo
			return nil
		}

		if err != syscall.EAGAIN && err != syscall.EWOULDBLOCK {
			_ = f.Close()
			return fmt.Errorf("flock failed: %w", err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func readAndCheckHolder(f *os.File) (*LockHolderInfo, error) {
	_, err := f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	holder, err := ParseLockHolderInfo(data)
	if err != nil {
		return nil, nil
	}

	if holder.PID == 0 {
		return nil, nil
	}

	if !holder.IsProcessAlive() {
		return holder, nil
	}

	return holder, nil
}

func writeHolderInfo(f *os.File, info *LockHolderInfo) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	if info == nil {
		return nil
	}

	data, err := info.ToBytes()
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func modeToString(mode LockMode) string {
	if mode == LockModeExclusive {
		return "exclusive"
	}
	return "shared"
}

func (l *Lock) GetLockFilePath() string {
	return l.lockFilePath
}

func (l *Lock) GetTargetFilePath() string {
	return l.filePath
}

func (l *Lock) IsHeld() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.holdCount > 0
}

func (l *Lock) Mode() LockMode {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.heldMode == nil {
		return LockModeExclusive
	}
	return *l.heldMode
}

func GetLockFilePathFor(target string) string {
	return filepath.Clean(target) + LockFileSuffix
}
