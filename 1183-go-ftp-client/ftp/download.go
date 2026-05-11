package ftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ProgressCallback func(transferred, total int64)

func (c *Client) DownloadFile(remotePath, localPath string, callback ProgressCallback) error {
	return c.DownloadFileWithContext(context.Background(), remotePath, localPath, callback)
}

func (c *Client) DownloadFileWithContext(ctx context.Context, remotePath, localPath string, callback ProgressCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	resolvedRemotePath := c.resolvePath(remotePath)

	totalSize, err := c.getSizeInternal(resolvedRemotePath)
	if err != nil {
		return err
	}

	offset := int64(0)
	if info, err := os.Stat(localPath); err == nil {
		offset = info.Size()
		if offset > totalSize {
			offset = 0
		}
	}

	if offset >= totalSize && totalSize > 0 {
		if callback != nil {
			callback(totalSize, totalSize)
		}
		return nil
	}

	dataConn, err := c.openDataConnection()
	if err != nil {
		return err
	}
	defer dataConn.Close()

	if offset > 0 {
		resp, err := c.sendCommand("REST %d", offset)
		if err != nil {
			return err
		}
		if resp.Code/100 != 3 {
			return fmt.Errorf("REST failed: %d %s", resp.Code, resp.Msg)
		}
	}

	resp, err := c.sendCommand("RETR %s", resolvedRemotePath)
	if err != nil {
		return err
	}
	if resp.Code/100 != 1 && resp.Code/100 != 2 {
		return fmt.Errorf("RETR failed: %d %s", resp.Code, resp.Msg)
	}

	var file *os.File
	if offset > 0 {
		file, err = os.OpenFile(localPath, os.O_WRONLY|os.O_APPEND, 0644)
	} else {
		dir := filepath.Dir(localPath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				dataConn.Close()
				return err
			}
		}
		file, err = os.Create(localPath)
	}
	if err != nil {
		dataConn.Close()
		return err
	}
	defer file.Close()

	buf := make([]byte, 64*1024)
	transferred := offset

	for {
		select {
		case <-ctx.Done():
			dataConn.Close()
			return ctx.Err()
		default:
		}

		n, readErr := dataConn.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				dataConn.Close()
				return writeErr
			}
			transferred += int64(n)
			if callback != nil {
				callback(transferred, totalSize)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	file.Sync()
	dataConn.Close()

	if err := c.readTransferComplete(); err != nil {
		return err
	}

	return nil
}

func (c *Client) DownloadDirectory(remoteDir, localDir string, callback ProgressCallback) error {
	return c.DownloadDirectoryWithContext(context.Background(), remoteDir, localDir, callback)
}

func (c *Client) DownloadDirectoryWithContext(ctx context.Context, remoteDir, localDir string, callback ProgressCallback) error {
	entries, err := c.List(remoteDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		remotePath := filepath.Join(remoteDir, entry.Name)
		localPath := filepath.Join(localDir, entry.Name)

		if entry.IsDirectory {
			if err := c.DownloadDirectoryWithContext(ctx, remotePath, localPath, callback); err != nil {
				return err
			}
		} else {
			if err := c.DownloadFileWithContext(ctx, remotePath, localPath, callback); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) getSizeInternal(path string) (int64, error) {
	resp, err := c.sendCommand("SIZE %s", path)
	if err != nil {
		return 0, err
	}
	if resp.Code/100 != 2 {
		return 0, nil
	}

	parts := strings.Fields(resp.Msg)
	if len(parts) == 0 {
		return 0, nil
	}

	var size int64
	if _, err := fmt.Sscanf(parts[0], "%d", &size); err != nil {
		return 0, nil
	}

	return size, nil
}
