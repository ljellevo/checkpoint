package checkpoint

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func checkpointsDir(gitDir string) string {
	return filepath.Join(gitDir, "checkpoints")
}

func ensureDir(gitDir string) error {
	return os.MkdirAll(checkpointsDir(gitDir), 0755)
}

func GenerateID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func LoadMeta(gitDir string) (*Meta, error) {
	path := filepath.Join(checkpointsDir(gitDir), "meta.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Meta{}, nil
	}
	if err != nil {
		return nil, err
	}
	var m Meta
	return &m, json.Unmarshal(data, &m)
}

func SaveMeta(gitDir string, m *Meta) error {
	if err := ensureDir(gitDir); err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(checkpointsDir(gitDir), "meta.json"), data, 0644)
}

func LoadCheckpoint(gitDir, id string) (*Checkpoint, error) {
	path := filepath.Join(checkpointsDir(gitDir), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cp Checkpoint
	return &cp, json.Unmarshal(data, &cp)
}

func SaveCheckpoint(gitDir string, cp *Checkpoint) error {
	if err := ensureDir(gitDir); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(checkpointsDir(gitDir), cp.ID+".json"), data, 0644)
}

func DeleteCheckpoint(gitDir, id string) error {
	return os.Remove(filepath.Join(checkpointsDir(gitDir), id+".json"))
}

// WalkChain returns checkpoints from HEAD down to the root (newest first).
func WalkChain(gitDir, headID string) ([]*Checkpoint, error) {
	if headID == "" {
		return nil, nil
	}
	var chain []*Checkpoint
	id := headID
	for id != "" {
		cp, err := LoadCheckpoint(gitDir, id)
		if err != nil {
			return nil, fmt.Errorf("loading checkpoint %s: %w", id, err)
		}
		chain = append(chain, cp)
		id = cp.ParentID
	}
	return chain, nil
}

// ListAll returns all checkpoint files in the checkpoints dir (unordered).
func ListAll(gitDir string) ([]*Checkpoint, error) {
	dir := checkpointsDir(gitDir)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var all []*Checkpoint
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || e.Name() == "meta.json" {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		cp, err := LoadCheckpoint(gitDir, id)
		if err != nil {
			return nil, err
		}
		all = append(all, cp)
	}
	return all, nil
}

// FindChild returns the checkpoint whose ParentID is parentID, or nil if none.
func FindChild(gitDir, parentID string) (*Checkpoint, error) {
	all, err := ListAll(gitDir)
	if err != nil {
		return nil, err
	}
	for _, cp := range all {
		if cp.ParentID == parentID {
			return cp, nil
		}
	}
	return nil, nil
}

// BuildFullChain returns the complete timeline (newest first), including future
// checkpoints that are still accessible beyond the current HEAD.
func BuildFullChain(gitDir, headID string) ([]*Checkpoint, error) {
	past, err := WalkChain(gitDir, headID)
	if err != nil {
		return nil, err
	}

	// Walk forward from HEAD to find any still-accessible future checkpoints.
	var future []*Checkpoint
	curID := headID
	for {
		child, err := FindChild(gitDir, curID)
		if err != nil {
			return nil, err
		}
		if child == nil {
			break
		}
		future = append(future, child)
		curID = child.ID
	}

	// Reverse future so the most recent is first.
	for i, j := 0, len(future)-1; i < j; i, j = i+1, j-1 {
		future[i], future[j] = future[j], future[i]
	}

	return append(future, past...), nil
}

// PruneOrphans deletes any checkpoint not reachable from currentHeadID.
func PruneOrphans(gitDir, currentHeadID string) error {
	chain, err := WalkChain(gitDir, currentHeadID)
	if err != nil {
		return err
	}
	reachable := make(map[string]bool, len(chain))
	for _, cp := range chain {
		reachable[cp.ID] = true
	}

	all, err := ListAll(gitDir)
	if err != nil {
		return err
	}
	for _, cp := range all {
		if !reachable[cp.ID] {
			if err := DeleteCheckpoint(gitDir, cp.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
