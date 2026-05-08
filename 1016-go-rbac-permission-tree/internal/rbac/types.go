package rbac

import "time"

type Scope string

const (
	ScopeSelf       Scope = "self"
	ScopeDepartment Scope = "department"
	ScopeAll        Scope = "all"
)

var scopeOrder = map[Scope]int{
	ScopeSelf:       0,
	ScopeDepartment: 1,
	ScopeAll:        2,
}

func (s Scope) Compare(other Scope) int {
	return scopeOrder[s] - scopeOrder[other]
}

func (s Scope) Max(other Scope) Scope {
	if s.Compare(other) >= 0 {
		return s
	}
	return other
}

type Permission struct {
	Resource string
	Action   string
	Scope    Scope
}

func (p Permission) Key() string {
	return p.Resource + ":" + p.Action
}

type Role struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
