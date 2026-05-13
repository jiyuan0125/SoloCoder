package diff

type Action string

const (
	ActionAdd    Action = "add"
	ActionRemove Action = "remove"
	ActionModify Action = "modify"
)

type Change struct {
	Path     string      `json:"path"`
	Action   Action      `json:"action"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
	Index    *int        `json:"index,omitempty"`
}
