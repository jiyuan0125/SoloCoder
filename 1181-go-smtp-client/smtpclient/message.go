package smtpclient

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"
	"unicode/utf8"
)

type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
	Date        time.Time
	MessageID   string
}

type Attachment struct {
	Filename    string
	Data        []byte
	ContentType string
}

func NewMessage() *Message {
	return &Message{
		Date:        time.Now(),
		Attachments: make([]Attachment, 0),
	}
}

func (m *Message) HasNonASCII() bool {
	if containsNonASCII(m.Subject) {
		return true
	}
	if containsNonASCII(m.TextBody) {
		return true
	}
	if containsNonASCII(m.HTMLBody) {
		return true
	}
	for _, att := range m.Attachments {
		if containsNonASCII(att.Filename) {
			return true
		}
	}
	return false
}

func containsNonASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
	}
	return false
}

func EncodeRFC2047(s string) string {
	if !containsNonASCII(s) {
		return s
	}
	
	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
}

func (m *Message) Build(supports8BITMIME bool) ([]byte, error) {
	var buf bytes.Buffer

	if m.MessageID == "" {
		m.MessageID = generateMessageID()
	}
	buf.WriteString(fmt.Sprintf("Message-ID: %s\r\n", m.MessageID))

	if m.From != "" {
		buf.WriteString(fmt.Sprintf("From: %s\r\n", m.From))
	}

	if len(m.To) > 0 {
		buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(m.To, ", ")))
	}

	if len(m.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(m.Cc, ", ")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", EncodeRFC2047(m.Subject)))
	
	dateStr := m.Date.Format(time.RFC1123Z)
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", dateStr))

	buf.WriteString("MIME-Version: 1.0\r\n")

	hasAttachments := len(m.Attachments) > 0
	hasHTML := m.HTMLBody != ""
	hasText := m.TextBody != ""

	if hasAttachments {
		mixedBoundary := generateBoundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", mixedBoundary))
		buf.WriteString("\r\n")

		if hasHTML || hasText {
			altBoundary := generateBoundary()
			buf.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
			buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", altBoundary))
			buf.WriteString("\r\n")

			if hasText {
				writeBodyPart(&buf, altBoundary, "text/plain", m.TextBody, supports8BITMIME)
			}
			if hasHTML {
				writeBodyPart(&buf, altBoundary, "text/html", m.HTMLBody, supports8BITMIME)
			}
			buf.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
			buf.WriteString("\r\n")
		}

		for _, att := range m.Attachments {
			writeAttachment(&buf, mixedBoundary, att)
		}
		buf.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))
	} else if hasHTML && hasText {
		altBoundary := generateBoundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", altBoundary))
		buf.WriteString("\r\n")
		writeBodyPart(&buf, altBoundary, "text/plain", m.TextBody, supports8BITMIME)
		writeBodyPart(&buf, altBoundary, "text/html", m.HTMLBody, supports8BITMIME)
		buf.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
	} else if hasHTML {
		writeSinglePart(&buf, "text/html", m.HTMLBody, supports8BITMIME)
	} else if hasText {
		writeSinglePart(&buf, "text/plain", m.TextBody, supports8BITMIME)
	}

	return buf.Bytes(), nil
}

func writeBodyPart(buf *bytes.Buffer, boundary, contentType, content string, supports8BITMIME bool) {
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n", contentType))
	
	if containsNonASCII(content) && !supports8BITMIME {
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		writeBase64Lines(buf, []byte(content))
	} else {
		buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(content)
		if !strings.HasSuffix(content, "\r\n") {
			buf.WriteString("\r\n")
		}
	}
	buf.WriteString("\r\n")
}

func writeSinglePart(buf *bytes.Buffer, contentType, content string, supports8BITMIME bool) {
	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n", contentType))
	
	if containsNonASCII(content) && !supports8BITMIME {
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("\r\n")
		writeBase64Lines(buf, []byte(content))
	} else {
		buf.WriteString("Content-Transfer-Encoding: 8bit\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(content)
		if !strings.HasSuffix(content, "\r\n") {
			buf.WriteString("\r\n")
		}
	}
}

func writeAttachment(buf *bytes.Buffer, boundary string, att Attachment) {
	ct := att.ContentType
	if ct == "" {
		ct = mime.TypeByExtension(att.Filename)
		if ct == "" {
			ct = "application/octet-stream"
		}
	}

	encodedFilename := EncodeRFC2047(att.Filename)

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s; name=%s\r\n", ct, encodedFilename))
	buf.WriteString("Content-Transfer-Encoding: base64\r\n")
	buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n", encodedFilename))
	buf.WriteString("\r\n")

	writeBase64Lines(buf, att.Data)
	buf.WriteString("\r\n")
}

func writeBase64Lines(w io.Writer, data []byte) {
	encoded := base64.StdEncoding.EncodeToString(data)
	const lineLen = 76
	for i := 0; i < len(encoded); i += lineLen {
		end := i + lineLen
		if end > len(encoded) {
			end = len(encoded)
		}
		w.Write([]byte(encoded[i:end]))
		w.Write([]byte("\r\n"))
	}
}

func generateBoundary() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func generateMessageID() string {
	b := make([]byte, 16)
	rand.Read(b)
	ts := time.Now().Unix()
	return fmt.Sprintf("<%d.%x@localhost>", ts, b)
}

func isUTF8(s string) bool {
	return utf8.ValidString(s)
}
