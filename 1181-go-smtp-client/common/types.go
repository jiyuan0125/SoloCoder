package common

type Attachment struct {
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
	ContentType string `json:"content_type,omitempty"`
}

type SendEmailRequest struct {
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Cc          []string     `json:"cc,omitempty"`
	Bcc         []string     `json:"bcc,omitempty"`
	Subject     string       `json:"subject"`
	TextBody    string       `json:"text_body,omitempty"`
	HTMLBody    string       `json:"html_body,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type QueueItem struct {
	ID          string       `json:"id"`
	Request     SendEmailRequest `json:"request"`
	Status      string       `json:"status"`
	CreatedAt   string       `json:"created_at"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

type SendEmailResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type GetQueueResponse struct {
	Items []QueueItem `json:"items"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
