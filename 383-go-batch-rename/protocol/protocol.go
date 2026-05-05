package protocol

import "encoding/json"

type MessageType string

const (
    MsgTypePreview       MessageType = "preview"
    MsgTypeExecute       MessageType = "execute"
    MsgTypeHistory       MessageType = "history"
    MsgTypePreviewResp   MessageType = "preview_resp"
    MsgTypeExecuteResp   MessageType = "execute_resp"
    MsgTypeHistoryResp   MessageType = "history_resp"
    MsgTypeError         MessageType = "error"
)

type RuleType string

const (
    RuleTypePrefix       RuleType = "prefix"
    RuleTypeSuffix       RuleType = "suffix"
    RuleTypeSequence     RuleType = "sequence"
    RuleTypeDate         RuleType = "date"
    RuleTypeReplace      RuleType = "replace"
    RuleTypeRegexReplace RuleType = "regex_replace"
    RuleTypeTemplate     RuleType = "template"
)

type Rule struct {
    Type        RuleType `json:"type"`
    Prefix      string   `json:"prefix,omitempty"`
    Suffix      string   `json:"suffix,omitempty"`
    DateFormat  string   `json:"date_format,omitempty"`
    OldString   string   `json:"old_string,omitempty"`
    NewString   string   `json:"new_string,omitempty"`
    Pattern     string   `json:"pattern,omitempty"`
    Replacement string   `json:"replacement,omitempty"`
    Template    string   `json:"template,omitempty"`
}

type Request struct {
    Type        MessageType `json:"type"`
    Files       []string    `json:"files,omitempty"`
    Rules       []Rule      `json:"rules,omitempty"`
    Recursive   bool        `json:"recursive,omitempty"`
    DryRun      bool        `json:"dry_run,omitempty"`
    DirPath     string      `json:"dir_path,omitempty"`
}

type RenameItem struct {
    OldPath   string `json:"old_path"`
    NewPath   string `json:"new_path"`
    OldName   string `json:"old_name"`
    NewName   string `json:"new_name"`
    Skipped   bool   `json:"skipped,omitempty"`
    SkipReason string `json:"skip_reason,omitempty"`
}

type PreviewResponse struct {
    Items     []RenameItem `json:"items"`
    Total     int          `json:"total"`
    Skipped   int          `json:"skipped"`
}

type ExecuteResponse struct {
    Items     []RenameItem `json:"items"`
    Success   int          `json:"success"`
    Failed    int          `json:"failed"`
    Skipped   int          `json:"skipped"`
}

type ErrorResponse struct {
    Message string `json:"message"`
}

type HistoryRecord struct {
    ID        int64        `json:"id"`
    Timestamp int64        `json:"timestamp"`
    Items     []RenameItem `json:"items"`
    Success   int          `json:"success"`
    Failed    int          `json:"failed"`
    Skipped   int          `json:"skipped"`
}

type HistoryResponse struct {
    Records []HistoryRecord `json:"records"`
    Total   int             `json:"total"`
}

func (r *Request) ToJSON() ([]byte, error) {
    return json.Marshal(r)
}

func (r *Request) FromJSON(data []byte) error {
    return json.Unmarshal(data, r)
}

func (r *PreviewResponse) ToJSON() ([]byte, error) {
    return json.Marshal(r)
}

func (r *PreviewResponse) FromJSON(data []byte) error {
    return json.Unmarshal(data, r)
}

func (r *ExecuteResponse) ToJSON() ([]byte, error) {
    return json.Marshal(r)
}

func (r *ExecuteResponse) FromJSON(data []byte) error {
    return json.Unmarshal(data, r)
}

func (r *ErrorResponse) ToJSON() ([]byte, error) {
    return json.Marshal(r)
}

func (r *ErrorResponse) FromJSON(data []byte) error {
    return json.Unmarshal(data, r)
}

func (r *HistoryResponse) ToJSON() ([]byte, error) {
    return json.Marshal(r)
}

func (r *HistoryResponse) FromJSON(data []byte) error {
    return json.Unmarshal(data, r)
}
