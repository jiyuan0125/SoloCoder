package ftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (c *Client) UploadFile(localPath, remotePath string, callback ProgressCallback) error {
	return c.UploadFileWithContext(context.Background(), localPath, remotePath, callback)
}

func (c *Client) UploadFileWithContext(ctx context.Context, localPath, remotePath string, callback ProgressCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	totalSize := fileInfo.Size()

	resolvedRemotePath := c.resolvePath(remotePath)

	offset := int64(0)
	remoteSize, err := c.getSizeInternal(resolvedRemotePath)
	if err == nil && remoteSize > 0 {
		if remoteSize < totalSize {
			offset = remoteSize
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

	var cmd string
	if offset > 0 {
		resp, err := c.sendCommand("REST %d", offset)
		if err != nil {
			dataConn.Close()
			return err
		}
		if resp.Code/100 != 3 {
			dataConn.Close()
			return fmt.Errorf("REST failed: %d %s", resp.Code, resp.Msg)
		}
		cmd = "STOR"
	} else {
		cmd = "STOR"
	}

	resp, err := c.sendCommand("%s %s", cmd, resolvedRemotePath)
	if err != nil {
		dataConn.Close()
		return err
	}
	if resp.Code/100 != 1 && resp.Code/100 != 2 {
		dataConn.Close()
		return fmt.Errorf("STOR failed: %d %s", resp.Code, resp.Msg)
	}

	file, err := os.Open(localPath)
	if err != nil {
		dataConn.Close()
		return err
	}
	defer file.Close()

	if offset > 0 {
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			dataConn.Close()
			return err
		}
	}

	buf := make([]byte, 64*1024)
	transferred := offset

	for {
		select {
		case <-ctx.Done():
			dataConn.Close()
			return ctx.Err()
		default:
		}

		n, readErr := file.Read(buf)
		if n > 0 {
			if _, writeErr := dataConn.Write(buf[:n]); writeErr != nil {
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

	dataConn.Close()

	if err := c.readTransferComplete(); err != nil {
		return err
	}

	return nil
}

func (c *Client) UploadDirectory(localDir, remoteDir string, callback ProgressCallback) error {
	return c.UploadDirectoryWithContext(context.Background(), localDir, remoteDir, callback)
}

func (c *Client) UploadDirectoryWithContext(ctx context.Context, localDir, remoteDir string, callback ProgressCallback) error {
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}

	if err := c.ensureRemoteDir(remoteDir); err != nil {
		return err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := entry.Name()
		localPath := filepath.Join(localDir, name)
		remotePath := remoteDir + "/" + name
		remotePath = cleanPath(remotePath)

		if entry.IsDir() {
			if err := c.UploadDirectoryWithContext(ctx, localPath, remotePath, callback); err != nil {
				return err
			}
		} else {
			if err := c.UploadFileWithContext(ctx, localPath, remotePath, callback); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) ensureRemoteDir(remoteDir string) error {
	parts := strings.Split(strings.Trim(remoteDir, "/"), "/")
	current := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		current = current + "/" + part
		current = cleanPath(current)

		resp, err := c.sendCommand("CWD %s", current)
		if err != nil {
			return err
		}
		if resp.Code/100 != 2 {
			if resp, err := c.sendCommand("MKD %s", current); err != nil {
				return err
			} else if resp.Code/100 != 2 {
				return fmt.Errorf("MKD failed: %d %s", resp.Code, resp.Msg)
			}
		}
	}

	return nil
}
