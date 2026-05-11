package sftp

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	sshConn   *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	version   uint32
	nextReqID uint32
	mu        sync.Mutex
}

type FileAttrs struct {
	Type      uint8
	Size      uint64
	UID       uint32
	GID       uint32
	Perms     uint32
	Atime     uint32
	Mtime     uint32
}

type File struct {
	client *Client
	handle []byte
	offset uint64
}

type DirEntry struct {
	Filename string
	Longname string
	Attrs    FileAttrs
}

type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	KeyPath     string
}

func NewClient(cfg *Config) (*Client, error) {
	var auth []ssh.AuthMethod

	if cfg.Password != "" {
		auth = append(auth, ssh.Password(cfg.Password))
	}

	if cfg.KeyPath != "" {
		key, err := os.ReadFile(cfg.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read key file: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	if len(auth) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	sshConn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	session, err := sshConn.NewSession()
	if err != nil {
		sshConn.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		sshConn.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		stdin.Close()
		session.Close()
		sshConn.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	err = session.RequestSubsystem("sftp")
	if err != nil {
		stdin.Close()
		session.Close()
		sshConn.Close()
		return nil, fmt.Errorf("failed to request sftp subsystem: %w", err)
	}

	client := &Client{
		sshConn:   sshConn,
		session:   session,
		stdin:     stdin,
		stdout:    stdout,
		nextReqID: 1,
	}

	if err := client.negotiateVersion(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

func (c *Client) negotiateVersion() error {
	clientVersion := uint32(3)

	pkt := make([]byte, 9)
	binary.BigEndian.PutUint32(pkt[0:4], 5)
	pkt[4] = SSH_FXP_INIT
	binary.BigEndian.PutUint32(pkt[5:9], clientVersion)

	if _, err := c.stdin.Write(pkt); err != nil {
		return fmt.Errorf("failed to send version init: %w", err)
	}

	respType, payload, err := c.readPacket()
	if err != nil {
		return fmt.Errorf("failed to read version response: %w", err)
	}

	if respType != SSH_FXP_VERSION {
		return fmt.Errorf("unexpected response type for version negotiation: %d", respType)
	}

	if len(payload) < 4 {
		return fmt.Errorf("version response too short")
	}

	serverVersion := binary.BigEndian.Uint32(payload[0:4])
	c.version = clientVersion
	if serverVersion < clientVersion {
		c.version = serverVersion
	}

	return nil
}

func (c *Client) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.session != nil {
		c.session.Close()
	}
	if c.sshConn != nil {
		return c.sshConn.Close()
	}
	return nil
}

func (c *Client) nextID() uint32 {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextReqID
	c.nextReqID++
	return id
}

func (c *Client) readPacket() (uint8, []byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.stdout, lenBuf); err != nil {
		return 0, nil, fmt.Errorf("failed to read packet length: %w", err)
	}

	pktLen := binary.BigEndian.Uint32(lenBuf)
	if pktLen == 0 {
		return 0, nil, fmt.Errorf("empty packet")
	}

	pktBuf := make([]byte, pktLen)
	if _, err := io.ReadFull(c.stdout, pktBuf); err != nil {
		return 0, nil, fmt.Errorf("failed to read packet: %w", err)
	}

	if pktLen < 1 {
		return 0, nil, fmt.Errorf("packet too short")
	}

	pktType := pktBuf[0]
	payload := pktBuf[1:]

	return pktType, payload, nil
}

func (c *Client) readResponse(expectedReqID uint32) (uint8, []byte, error) {
	for {
		pktType, payload, err := c.readPacket()
		if err != nil {
			return 0, nil, err
		}

		if pktType == SSH_FXP_VERSION {
			continue
		}

		if len(payload) < 4 {
			return 0, nil, fmt.Errorf("response packet too short")
		}

		reqID := binary.BigEndian.Uint32(payload[0:4])
		if reqID == expectedReqID {
			return pktType, payload[4:], nil
		}
	}
}

func (c *Client) sendPacket(pktType uint8, payload []byte) (uint32, error) {
	reqID := c.nextID()

	pktLen := 1 + 4 + len(payload)
	pkt := make([]byte, 4+pktLen)

	binary.BigEndian.PutUint32(pkt[0:4], uint32(pktLen))
	pkt[4] = pktType
	binary.BigEndian.PutUint32(pkt[5:9], reqID)
	copy(pkt[9:], payload)

	if _, err := c.stdin.Write(pkt); err != nil {
		return 0, fmt.Errorf("failed to send packet: %w", err)
	}

	return reqID, nil
}

func (c *Client) OpenFile(path string, flags uint32) (*File, error) {
	pathBytes := []byte(path)
	payload := make([]byte, 4+len(pathBytes)+4+4)

	binary.BigEndian.PutUint32(payload[0:4], uint32(len(pathBytes)))
	copy(payload[4:4+len(pathBytes)], pathBytes)
	offset := 4 + len(pathBytes)
	binary.BigEndian.PutUint32(payload[offset:offset+4], flags)
	offset += 4
	binary.BigEndian.PutUint32(payload[offset:offset+4], 0)

	reqID, err := c.sendPacket(SSH_FXP_OPEN, payload)
	if err != nil {
		return nil, err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return nil, err
	}

	if pktType == SSH_FXP_STATUS {
		return nil, parseStatusError(resp)
	}

	if pktType != SSH_FXP_HANDLE {
		return nil, fmt.Errorf("unexpected response type: %d", pktType)
	}

	handle, _, err := readString(resp)
	if err != nil {
		return nil, err
	}

	return &File{
		client: c,
		handle: handle,
	}, nil
}

func (c *Client) CloseFile(f *File) error {
	payload := make([]byte, 4+len(f.handle))
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(f.handle)))
	copy(payload[4:], f.handle)

	reqID, err := c.sendPacket(SSH_FXP_CLOSE, payload)
	if err != nil {
		return err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return err
	}

	if pktType != SSH_FXP_STATUS {
		return fmt.Errorf("unexpected response type: %d", pktType)
	}

	return parseStatusError(resp)
}

func (c *Client) ReadFile(f *File, offset uint64, length uint32) ([]byte, error) {
	payload := make([]byte, 4+len(f.handle)+8+4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(f.handle)))
	copy(payload[4:], f.handle)
	off := 4 + len(f.handle)
	binary.BigEndian.PutUint64(payload[off:off+8], offset)
	off += 8
	binary.BigEndian.PutUint32(payload[off:off+4], length)

	reqID, err := c.sendPacket(SSH_FXP_READ, payload)
	if err != nil {
		return nil, err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return nil, err
	}

	if pktType == SSH_FXP_STATUS {
		err := parseStatusError(resp)
		if statusErr, ok := err.(*StatusError); ok && statusErr.Code == SSH_FX_EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	if pktType != SSH_FXP_DATA {
		return nil, fmt.Errorf("unexpected response type: %d", pktType)
	}

	data, _, err := readString(resp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Client) WriteFile(f *File, offset uint64, data []byte) error {
	payload := make([]byte, 4+len(f.handle)+8+4+len(data))
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(f.handle)))
	copy(payload[4:], f.handle)
	off := 4 + len(f.handle)
	binary.BigEndian.PutUint64(payload[off:off+8], offset)
	off += 8
	binary.BigEndian.PutUint32(payload[off:off+4], uint32(len(data)))
	off += 4
	copy(payload[off:], data)

	reqID, err := c.sendPacket(SSH_FXP_WRITE, payload)
	if err != nil {
		return err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return err
	}

	if pktType != SSH_FXP_STATUS {
		return fmt.Errorf("unexpected response type: %d", pktType)
	}

	return parseStatusError(resp)
}

func (c *Client) OpenDir(path string) (*File, error) {
	pathBytes := []byte(path)
	payload := make([]byte, 4+len(pathBytes))
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(pathBytes)))
	copy(payload[4:], pathBytes)

	reqID, err := c.sendPacket(SSH_FXP_OPENDIR, payload)
	if err != nil {
		return nil, err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return nil, err
	}

	if pktType == SSH_FXP_STATUS {
		return nil, parseStatusError(resp)
	}

	if pktType != SSH_FXP_HANDLE {
		return nil, fmt.Errorf("unexpected response type: %d", pktType)
	}

	handle, _, err := readString(resp)
	if err != nil {
		return nil, err
	}

	return &File{
		client: c,
		handle: handle,
	}, nil
}

func (c *Client) ReadDir(f *File) ([]DirEntry, error) {
	payload := make([]byte, 4+len(f.handle))
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(f.handle)))
	copy(payload[4:], f.handle)

	reqID, err := c.sendPacket(SSH_FXP_READDIR, payload)
	if err != nil {
		return nil, err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return nil, err
	}

	if pktType == SSH_FXP_STATUS {
		statusErr := parseStatusError(resp)
		if sErr, ok := statusErr.(*StatusError); ok && sErr.Code == SSH_FX_EOF {
			return nil, io.EOF
		}
		return nil, statusErr
	}

	if pktType != SSH_FXP_NAME {
		return nil, fmt.Errorf("unexpected response type: %d", pktType)
	}

	entries, err := parseDirEntries(resp)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, io.EOF
	}

	return entries, nil
}

func (c *Client) ListDir(path string) ([]DirEntry, error) {
	dir, err := c.OpenDir(path)
	if err != nil {
		return nil, err
	}
	defer c.CloseFile(dir)

	var allEntries []DirEntry
	for {
		entries, err := c.ReadDir(dir)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

func (c *Client) Stat(path string) (*FileAttrs, error) {
	pathBytes := []byte(path)
	payload := make([]byte, 4+len(pathBytes))
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(pathBytes)))
	copy(payload[4:], pathBytes)

	var pktType uint8
	var resp []byte
	var sendErr error

	if c.version >= 3 {
		reqID, err := c.sendPacket(SSH_FXP_STAT_V3, payload)
		if err != nil {
			return nil, err
		}
		pktType, resp, sendErr = c.readResponse(reqID)
	} else {
		reqID, err := c.sendPacket(SSH_FXP_STAT, payload)
		if err != nil {
			return nil, err
		}
		pktType, resp, sendErr = c.readResponse(reqID)
	}

	if sendErr != nil {
		return nil, sendErr
	}

	if pktType == SSH_FXP_STATUS {
		return nil, parseStatusError(resp)
	}

	if pktType != SSH_FXP_ATTRS {
		return nil, fmt.Errorf("unexpected response type: %d", pktType)
	}

	return parseAttrs(resp, c.version)
}

func (c *Client) Fstat(f *File) (*FileAttrs, error) {
	payload := make([]byte, 4+len(f.handle))
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(f.handle)))
	copy(payload[4:], f.handle)

	reqID, err := c.sendPacket(SSH_FXP_FSTAT, payload)
	if err != nil {
		return nil, err
	}

	pktType, resp, err := c.readResponse(reqID)
	if err != nil {
		return nil, err
	}

	if pktType == SSH_FXP_STATUS {
		return nil, parseStatusError(resp)
	}

	if pktType != SSH_FXP_ATTRS {
		return nil, fmt.Errorf("unexpected response type: %d", pktType)
	}

	return parseAttrs(resp, c.version)
}

func (c *Client) UploadFile(localPath, remotePath string, progress chan<- int64) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	info, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	flags := uint32(SSH_FXF_WRITE | SSH_FXF_CREATE | SSH_FXF_TRUNCATE)
	remoteFile, err := c.OpenFile(remotePath, flags)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer c.CloseFile(remoteFile)

	var totalWritten int64
	buf := make([]byte, 32*1024)
	var offset uint64 = 0

	for {
		n, err := localFile.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read local file: %w", err)
		}

		if err := c.WriteFile(remoteFile, offset, buf[:n]); err != nil {
			return fmt.Errorf("failed to write to remote file: %w", err)
		}

		offset += uint64(n)
		totalWritten += int64(n)

		if progress != nil {
			select {
			case progress <- totalWritten:
			default:
			}
		}
	}

	_ = info
	return nil
}

func (c *Client) DownloadFile(remotePath, localPath string, progress chan<- int64) error {
	flags := uint32(SSH_FXF_READ)
	remoteFile, err := c.OpenFile(remotePath, flags)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer c.CloseFile(remoteFile)

	attrs, err := c.Fstat(remoteFile)
	if err != nil {
		return fmt.Errorf("failed to get remote file info: %w", err)
	}
	_ = attrs

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	var totalRead int64
	var offset uint64 = 0
	bufSize := uint32(32 * 1024)

	for {
		data, err := c.ReadFile(remoteFile, offset, bufSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read remote file: %w", err)
		}

		if len(data) == 0 {
			break
		}

		if _, err := localFile.Write(data); err != nil {
			return fmt.Errorf("failed to write to local file: %w", err)
		}

		offset += uint64(len(data))
		totalRead += int64(len(data))

		if progress != nil {
			select {
			case progress <- totalRead:
			default:
			}
		}
	}

	return nil
}

func (f *File) IsDir() bool {
	return f != nil
}

func readString(data []byte) ([]byte, []byte, error) {
	if len(data) < 4 {
		return nil, nil, fmt.Errorf("data too short for string length")
	}

	strLen := binary.BigEndian.Uint32(data[0:4])
	if uint32(len(data)-4) < strLen {
		return nil, nil, fmt.Errorf("data too short for string content")
	}

	return data[4 : 4+strLen], data[4+strLen:], nil
}

func parseStatusError(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("status response too short")
	}

	code := binary.BigEndian.Uint32(data[0:4])

	if code == SSH_FX_OK {
		return nil
	}

	msg := "unknown error"
	if len(data) >= 8 {
		msgBytes, _, err := readString(data[4:])
		if err == nil {
			msg = string(msgBytes)
		}
	}

	return &StatusError{
		Code:    code,
		Message: msg,
	}
}

type StatusError struct {
	Code    uint32
	Message string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("sftp error %d: %s", e.Code, e.Message)
}

func parseAttrs(data []byte, version uint32) (*FileAttrs, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("attrs data too short")
	}

	flags := binary.BigEndian.Uint32(data[0:4])
	data = data[4:]

	attrs := &FileAttrs{
		Type: SSH_FILEXFER_TYPE_REGULAR,
	}

	if flags&SSH_FILEXFER_ATTR_SIZE != 0 {
		if len(data) < 8 {
			return nil, fmt.Errorf("attrs data too short for size")
		}
		attrs.Size = binary.BigEndian.Uint64(data[0:8])
		data = data[8:]
	}

	if flags&SSH_FILEXFER_ATTR_UIDGID != 0 {
		if len(data) < 8 {
			return nil, fmt.Errorf("attrs data too short for uid/gid")
		}
		attrs.UID = binary.BigEndian.Uint32(data[0:4])
		attrs.GID = binary.BigEndian.Uint32(data[4:8])
		data = data[8:]
	}

	if flags&SSH_FILEXFER_ATTR_PERMISSIONS != 0 {
		if version >= 4 {
			if len(data) < 4 {
				return nil, fmt.Errorf("attrs data too short for perms")
			}
			attrs.Perms = binary.BigEndian.Uint32(data[0:4])
			data = data[4:]
		} else {
			if len(data) < 4 {
				return nil, fmt.Errorf("attrs data too short for perms")
			}
			attrs.Perms = binary.BigEndian.Uint32(data[0:4])
			data = data[4:]
		}

		if attrs.Perms&040000 != 0 {
			attrs.Type = SSH_FILEXFER_TYPE_DIRECTORY
		} else if attrs.Perms&0120000 != 0 {
			attrs.Type = SSH_FILEXFER_TYPE_SYMLINK
		}
	}

	if flags&SSH_FILEXFER_ATTR_ACMODTIME != 0 {
		if len(data) < 8 {
			return nil, fmt.Errorf("attrs data too short for atime/mtime")
		}
		attrs.Atime = binary.BigEndian.Uint32(data[0:4])
		attrs.Mtime = binary.BigEndian.Uint32(data[4:8])
		data = data[8:]
	}

	if flags&SSH_FILEXFER_ATTR_EXTENDED != 0 {
		if len(data) < 4 {
			return nil, fmt.Errorf("attrs data too short for extended")
		}
		extCount := binary.BigEndian.Uint32(data[0:4])
		data = data[4:]
		for i := uint32(0); i < extCount; i++ {
			if len(data) < 4 {
				return nil, fmt.Errorf("attrs data too short for ext name")
			}
			nameLen := binary.BigEndian.Uint32(data[0:4])
			data = data[4:]
			if uint32(len(data)) < nameLen {
				return nil, fmt.Errorf("attrs data too short for ext name content")
			}
			data = data[nameLen:]

			if len(data) < 4 {
				return nil, fmt.Errorf("attrs data too short for ext data")
			}
			dataLen := binary.BigEndian.Uint32(data[0:4])
			data = data[4:]
			if uint32(len(data)) < dataLen {
				return nil, fmt.Errorf("attrs data too short for ext data content")
			}
			data = data[dataLen:]
		}
	}

	return attrs, nil
}

func parseDirEntries(data []byte) ([]DirEntry, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("dir entries data too short")
	}

	count := binary.BigEndian.Uint32(data[0:4])
	data = data[4:]

	entries := make([]DirEntry, 0, count)

	for i := uint32(0); i < count; i++ {
		filename, rest, err := readString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read filename: %w", err)
		}
		data = rest

		longname, rest, err := readString(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read longname: %w", err)
		}
		data = rest

		attrs, err := parseAttrs(data, 3)
		if err != nil {
			return nil, fmt.Errorf("failed to parse attrs: %w", err)
		}

		entry := DirEntry{
			Filename: string(filename),
			Longname: string(longname),
			Attrs:    *attrs,
		}
		entries = append(entries, entry)

		if strings.HasPrefix(entry.Longname, "d") {
			entry.Attrs.Type = SSH_FILEXFER_TYPE_DIRECTORY
		} else if strings.HasPrefix(entry.Longname, "l") {
			entry.Attrs.Type = SSH_FILEXFER_TYPE_SYMLINK
		}
	}

	return entries, nil
}
