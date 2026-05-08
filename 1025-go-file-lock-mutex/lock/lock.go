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
	}
}

func (l *Lock) LockExclusive(timeout time.Duration) error {
	return l.acquire(LockModeExclusive, timeout)
}

func (l *Lock) LockShared(timeout time.Duration) error {
	return l.acquire(LockModeShared, timeout)
}

func (l *Lock) Unlock() error {
	l.holdMu.Lock()
	defer l.holdMu.Unlock()

	if l.holdCount <= 0 {
		return ErrLockNotHeld
	}

	l.holdCount--
	if l.holdCount > 0 {
		if l.mode == LockModeExclusive {
			l.mu.Unlock()
		} else {
			l.localMu.RUnlock()
		}
		return nil
	}

	if l.file == nil {
		l.held = false
		return nil
	}

	if err := writeHolderInfo(l.file, nil); err != nil {
		_ = l.file.Close()
		l.file = nil
		l.held = false
		l.mode = LockModeExclusive
		return err
	}

	if err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN); err != nil {
		_ = l.file.Close()
		l.file = nil
		l.held = false
		l.mode = LockModeExclusive
		return err
	}

	if err := l.file.Close(); err != nil {
		l.file = nil
		l.held = false
		l.mode = LockModeExclusive
		return err
	}

	l.file = nil
	l.held = false
	l.mode = LockModeExclusive
	return nil
}

func (l *Lock) acquire(mode LockMode, timeout time.Duration) error {
	config := DefaultConfig()

	actualTimeout := timeout
	if timeout <= 0 {
		actualTimeout = config.Timeout
	}

	if mode != LockModeExclusive && mode != LockModeShared {
		return ErrInvalidMode
	}

	l.holdMu.Lock()
	if l.holdCount > 0 {
		if l.mode == LockModeExclusive || l.mode == mode {
			l.holdCount++
			l.holdMu.Unlock()
			if l.mode == LockModeExclusive {
				l.mu.Lock()
				l.mu.Unlock()
			} else {
				l.localMu.RLock()
				l.localMu.RUnlock()
			}
			return nil
		}
		l.holdMu.Unlock()
		return ErrInvalidMode
	}
	l.holdMu.Unlock()

	if mode == LockModeExclusive {
		l.mu.Lock()
	} else {
		l.localMu.RLock()
	}

	if err := l.doAcquire(mode, actualTimeout, config.MaxWaitDuration); err != nil {
		if mode == LockModeExclusive {
			l.mu.Unlock()
		} else {
			l.localMu.RUnlock()
		}
		return err
	}

	l.holdMu.Lock()
	l.holdCount = 1
	l.holdMu.Unlock()

	return nil
}

func (l *Lock) doAcquire(mode LockMode, timeout, maxWait time.Duration) error {
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
		elapsed := time.Since(startTime)
		if elapsed > maxWait {
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

			l.file = f
			l.mode = mode
			l.held = true
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
	l.holdMu.Lock()
	defer l.holdMu.Unlock()
	return l.holdCount > 0
}

func (l *Lock) Mode() LockMode {
	return l.mode
}

func GetLockFilePathFor(target string) string {
	return filepath.Clean(target) + LockFileSuffix
}
