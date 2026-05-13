package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"email-template/internal/config"
	"email-template/internal/database"
	"email-template/internal/template"
)

var retryIntervals = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

const maxRetries = 3
const maxBatchSize = 1000

type InlineImage struct {
	CID      string
	Filename string
	Data     []byte
}

type SendRequest struct {
	TemplateName  string
	Recipient     string
	Variables     map[string]interface{}
	InlineImages  []InlineImage
	BatchID       string
}

type BatchSendRequest struct {
	TemplateName  string
	CSVData       []byte
	Variables     map[string]interface{}
	InlineImages  []InlineImage
}

type SendResponse struct {
	RecordID   string
	Status     string
	Error      string
	Recipient  string
}

type BatchSendResponse struct {
	BatchID        string
	SuccessCount   int
	FailedCount    int
	SkippedEmails  []string
	Records        []SendResponse
}

type Service struct {
	cfg      *config.Config
	db       *database.DB
	tplSvc   *template.Service
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewService(cfg *config.Config, db *database.DB, tplSvc *template.Service) *Service {
	return &Service{
		cfg:      cfg,
		db:       db,
		tplSvc:   tplSvc,
		stopChan: make(chan struct{}),
	}
}

func (s *Service) StartRetryWorker() {
	s.wg.Add(1)
	go s.retryLoop()
}

func (s *Service) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Service) retryLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processPendingRecords()
		}
	}
}

func (s *Service) processPendingRecords() {
	ctx := context.Background()
	records, err := s.db.GetPendingRecordsForRetry(ctx, 50)
	if err != nil {
		return
	}

	for _, record := range records {
		go func(r *database.SendRecord) {
			s.retryRecord(ctx, r)
		}(record)
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Service) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	tpl, err := s.tplSvc.Get(ctx, req.TemplateName)
	if err != nil {
		return nil, err
	}

	recordID := generateID()
	variablesJSON, _ := json.Marshal(req.Variables)

	record := &database.SendRecord{
		ID:           recordID,
		TemplateID:   tpl.ID,
		TemplateName: tpl.Name,
		Recipient:    req.Recipient,
		Subject:      tpl.Subject,
		Variables:    string(variablesJSON),
		Status:       "pending",
		Attempts:     0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		BatchID:      req.BatchID,
	}

	if err := s.db.CreateSendRecord(ctx, record); err != nil {
		return nil, err
	}

	go s.sendNow(ctx, record, tpl, req.Variables, req.InlineImages)

	return &SendResponse{
		RecordID:  recordID,
		Status:    "queued",
		Recipient: req.Recipient,
	}, nil
}

func (s *Service) sendNow(ctx context.Context, record *database.SendRecord, tpl *database.Template,
	variables map[string]interface{}, inlineImages []InlineImage) {

	var vars map[string]interface{}
	if record.Variables != "" {
		json.Unmarshal([]byte(record.Variables), &vars)
	} else {
		vars = variables
	}

	htmlContent, err := s.tplSvc.RenderTemplate(ctx, tpl.HTML, vars)
	if err != nil {
		record.LastError = fmt.Sprintf("render failed: %v", err)
		record.Status = "failed"
		s.db.UpdateSendRecord(ctx, record)
		return
	}

	subjectContent, err := s.tplSvc.RenderTemplate(ctx, tpl.Subject, vars)
	if err != nil {
		subjectContent = tpl.Subject
	}

	record.Attempts++

	err = s.sendEmail(record.Recipient, subjectContent, htmlContent, inlineImages)
	if err == nil {
		record.Status = "sent"
		record.LastError = ""
		record.NextRetryAt = nil
		record.SendResult = "success"
		s.db.UpdateSendRecord(ctx, record)
		return
	}

	record.LastError = err.Error()
	if record.Attempts >= maxRetries {
		record.Status = "failed"
		record.NextRetryAt = nil
	} else {
		record.Status = "retry"
		nextRetry := time.Now().Add(retryIntervals[record.Attempts-1])
		record.NextRetryAt = &nextRetry
	}

	s.db.UpdateSendRecord(ctx, record)
}

func (s *Service) retryRecord(ctx context.Context, record *database.SendRecord) {
	tpl, err := s.db.GetTemplateByName(ctx, record.TemplateName)
	if err != nil {
		return
	}
	if tpl == nil {
		record.Status = "failed"
		record.LastError = "template not found"
		s.db.UpdateSendRecord(ctx, record)
		return
	}

	s.sendNow(ctx, record, tpl, nil, nil)
}

func (s *Service) sendEmail(to, subject, htmlBody string, inlineImages []InlineImage) error {
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	from := s.cfg.SMTPFrom
	if from == "" {
		from = s.cfg.SMTPUsername
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	boundary := writer.Boundary()

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/related; boundary=%s\r\n\r\n",
		from, to, encodeSubject(subject), boundary)
	body.WriteString(headers)

	htmlPart, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=utf-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return err
	}
	htmlPart.Write(encodeQuotedPrintable(htmlBody))

	for _, img := range inlineImages {
		imgPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {getMimeType(img.Filename)},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {fmt.Sprintf("inline; filename=\"%s\"", img.Filename)},
			"Content-ID":                {fmt.Sprintf("<%s>", img.CID)},
		})
		if err != nil {
			continue
		}
		encoder := base64.NewEncoder(base64.StdEncoding, imgPart)
		encoder.Write(img.Data)
		encoder.Close()
	}

	writer.Close()

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	return smtp.SendMail(addr, auth, from, []string{to}, body.Bytes())
}

func encodeSubject(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}

func encodeQuotedPrintable(s string) []byte {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '=' || b == '\t' || b == ' ' || b < 32 || b > 126 {
			fmt.Fprintf(&buf, "=%02X", b)
		} else {
			buf.WriteByte(b)
		}
	}
	return buf.Bytes()
}

func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func (s *Service) BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error) {
	emails, skipped, err := s.parseCSVEmails(req.CSVData)
	if err != nil {
		return nil, err
	}

	if len(emails) > maxBatchSize {
		return nil, fmt.Errorf("batch size exceeds %d", maxBatchSize)
	}

	batchID := generateID()
	response := &BatchSendResponse{
		BatchID:       batchID,
		SkippedEmails: skipped,
	}

	for _, email := range emails {
		sendReq := &SendRequest{
			TemplateName: req.TemplateName,
			Recipient:    email,
			Variables:    req.Variables,
			InlineImages: req.InlineImages,
			BatchID:      batchID,
		}

		resp, sendErr := s.Send(ctx, sendReq)
		if sendErr != nil {
			response.FailedCount++
			response.Records = append(response.Records, SendResponse{
				Recipient: email,
				Status:    "error",
				Error:     sendErr.Error(),
			})
		} else {
			response.SuccessCount++
			response.Records = append(response.Records, *resp)
		}
	}

	return response, nil
}

func (s *Service) parseCSVEmails(data []byte) ([]string, []string, error) {
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("empty CSV data")
	}

	lines := bytes.Split(data, []byte("\n"))
	var emails []string
	var skipped []string
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		parts := bytes.Split(line, []byte(","))
		email := string(bytes.TrimSpace(parts[0]))

		if email == "" {
			continue
		}

		if !emailRegex.MatchString(email) {
			skipped = append(skipped, email)
			continue
		}

		emails = append(emails, email)
	}

	return emails, skipped, nil
}

func (s *Service) GetSendStatus(ctx context.Context, id string) (*database.SendRecord, error) {
	record, err := s.db.GetSendRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("not found")
	}
	return record, nil
}

func (s *Service) GetBatchStatus(ctx context.Context, batchID string) ([]*database.SendRecord, error) {
	return s.db.GetRecordsByBatch(ctx, batchID)
}

func (s *Service) UploadImageFromURL(ctx context.Context, url string) (*InlineImage, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(url)
	cid := generateID()

	return &InlineImage{
		CID:      cid,
		Filename: filename,
		Data:     data,
	}, nil
}
