package checkpoint

import "time"

type Checkpoint struct {
	ID        string            `json:"id"`
	ParentID  string            `json:"parent_id,omitempty"`
	Message   string            `json:"message,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Patch     string            `json:"patch"`
	Untracked map[string]string `json:"untracked,omitempty"`
}

type Meta struct {
	Head string `json:"head,omitempty"`
}
