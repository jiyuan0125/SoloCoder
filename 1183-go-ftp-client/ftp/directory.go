package ftp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type FileEntry struct {
	Name        string
	IsDirectory bool
	Size        int64
	Modified    time.Time
	Permissions string
}

func (c *Client) List(path string) ([]FileEntry, error) {
	return c.listInternal(path, false)
}

func (c *Client) MLSD(path string) ([]FileEntry, error) {
	return c.listInternal(path, true)
}

func (c *Client) listInternal(path string, useMLSD bool) ([]FileEntry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	resolvedPath := c.resolvePath(path)

	dataConn, err := c.openDataConnection()
	if err != nil {
		return nil, err
	}
	defer dataConn.Close()

	var cmd string
	if useMLSD {
		if resolvedPath == "" || resolvedPath == "." {
			cmd = "MLSD"
		} else {
			cmd = fmt.Sprintf("MLSD %s", resolvedPath)
		}
	} else {
		if resolvedPath == "" || resolvedPath == "." {
			cmd = "LIST"
		} else {
			cmd = fmt.Sprintf("LIST %s", resolvedPath)
		}
	}

	resp, err := c.sendCommand(cmd)
	if err != nil {
		return nil, err
	}
	if resp.Code/100 != 1 && resp.Code/100 != 2 {
		return nil, fmt.Errorf("%s failed: %d %s", cmd, resp.Code, resp.Msg)
	}

	var entries []FileEntry
	reader := bufio.NewReader(dataConn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		var entry *FileEntry
		if useMLSD {
			entry = parseMLSDEntry(line)
		} else {
			entry = parseLISTEntry(line)
		}
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	dataConn.Close()

	if err := c.readTransferComplete(); err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *Client) ChangeDir(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	resolvedPath := c.resolvePath(path)

	resp, err := c.sendCommand("CWD %s", resolvedPath)
	if err != nil {
		return err
	}
	if resp.Code/100 != 2 {
		return fmt.Errorf("CWD failed: %d %s", resp.Code, resp.Msg)
	}

	c.currentDir = resolvedPath
	return nil
}

func (c *Client) MakeDir(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	resolvedPath := c.resolvePath(path)

	resp, err := c.sendCommand("MKD %s", resolvedPath)
	if err != nil {
		return err
	}
	if resp.Code/100 != 2 {
		return fmt.Errorf("MKD failed: %d %s", resp.Code, resp.Msg)
	}

	return nil
}

func (c *Client) RemoveDir(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	resolvedPath := c.resolvePath(path)

	resp, err := c.sendCommand("RMD %s", resolvedPath)
	if err != nil {
		return err
	}
	if resp.Code/100 != 2 {
		return fmt.Errorf("RMD failed: %d %s", resp.Code, resp.Msg)
	}

	return nil
}

func (c *Client) DeleteFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return err
	}

	resolvedPath := c.resolvePath(path)

	resp, err := c.sendCommand("DELE %s", resolvedPath)
	if err != nil {
		return err
	}
	if resp.Code/100 != 2 {
		return fmt.Errorf("DELE failed: %d %s", resp.Code, resp.Msg)
	}

	return nil
}

func (c *Client) GetSize(path string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return 0, err
	}

	resolvedPath := c.resolvePath(path)

	resp, err := c.sendCommand("SIZE %s", resolvedPath)
	if err != nil {
		return 0, err
	}
	if resp.Code/100 != 2 {
		return 0, fmt.Errorf("SIZE failed: %d %s", resp.Code, resp.Msg)
	}

	parts := strings.Fields(resp.Msg)
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid SIZE response: %s", resp.Msg)
	}

	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %s", parts[0])
	}

	return size, nil
}

func parseMLSDEntry(line string) *FileEntry {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil
	}

	factsStr := parts[0]
	name := parts[1]

	facts := make(map[string]string)
	factParts := strings.Split(factsStr, ";")
	for _, f := range factParts {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		eqIdx := strings.Index(f, "=")
		if eqIdx > 0 {
			key := strings.ToLower(f[:eqIdx])
			value := f[eqIdx+1:]
			facts[key] = value
		}
	}

	entry := &FileEntry{Name: name}

	if typeFact, ok := facts["type"]; ok {
		entry.IsDirectory = strings.EqualFold(typeFact, "dir") || strings.EqualFold(typeFact, "cdir") || strings.EqualFold(typeFact, "pdir")
	}

	if sizeFact, ok := facts["size"]; ok {
		if size, err := strconv.ParseInt(sizeFact, 10, 64); err == nil {
			entry.Size = size
		}
	}

	if modifyFact, ok := facts["modify"]; ok {
		if t, err := parseMLSDTime(modifyFact); err == nil {
			entry.Modified = t
		}
	}

	if permFact, ok := facts["perm"]; ok {
		entry.Permissions = permFact
	}

	return entry
}

func parseMLSDTime(s string) (time.Time, error) {
	formats := []string{
		"20060102150405",
		"20060102150405.000",
		"200601021504",
		"20060102",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %s", s)
}

func parseLISTEntry(line string) *FileEntry {
	if line == "" {
		return nil
	}

	if strings.HasPrefix(line, "total") {
		return nil
	}

	fields := strings.Fields(line)
	if len(fields) < 9 {
		return nil
	}

	entry := &FileEntry{}

	perm := fields[0]
	entry.Permissions = perm
	entry.IsDirectory = strings.HasPrefix(perm, "d")

	if sizeStr := fields[4]; sizeStr != "" {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			entry.Size = size
		}
	}

	month := fields[5]
	day := fields[6]
	yearOrTime := fields[7]
	t, err := parseLISTTime(month, day, yearOrTime)
	if err == nil {
		entry.Modified = t
	}

	entry.Name = strings.Join(fields[8:], " ")

	return entry
}

func parseLISTTime(month, day, yearOrTime string) (time.Time, error) {
	months := map[string]time.Month{
		"Jan": time.January, "Feb": time.February, "Mar": time.March,
		"Apr": time.April, "May": time.May, "Jun": time.June,
		"Jul": time.July, "Aug": time.August, "Sep": time.September,
		"Oct": time.October, "Nov": time.November, "Dec": time.December,
	}

	m, ok := months[month]
	if !ok {
		return time.Time{}, fmt.Errorf("invalid month: %s", month)
	}

	d, err := strconv.Atoi(day)
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	year := now.Year()

	if strings.Contains(yearOrTime, ":") {
		timeParts := strings.Split(yearOrTime, ":")
		if len(timeParts) == 2 {
			hour, _ := strconv.Atoi(timeParts[0])
			minute, _ := strconv.Atoi(timeParts[1])
			t := time.Date(year, m, d, hour, minute, 0, 0, time.Local)
			if t.After(now.AddDate(0, 0, 1)) {
				t = t.AddDate(-1, 0, 0)
			}
			return t, nil
		}
	} else {
		y, err := strconv.Atoi(yearOrTime)
		if err != nil {
			return time.Time{}, err
		}
		year = y
	}

	return time.Date(year, m, d, 0, 0, 0, 0, time.Local), nil
}
