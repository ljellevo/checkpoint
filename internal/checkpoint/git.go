package checkpoint

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetGitDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func hasHEAD() bool {
	return exec.Command("git", "rev-parse", "--verify", "HEAD").Run() == nil
}

// CaptureDiff returns the working-tree diff against HEAD.
// When there are no commits yet, staged files are stored via CaptureUntracked
// instead, so this returns an empty string.
func CaptureDiff() (string, error) {
	if !hasHEAD() {
		return "", nil
	}
	out, err := exec.Command("git", "diff", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("capturing diff: %w", err)
	}
	return string(out), nil
}

func CaptureUntracked() (map[string]string, error) {
	result := map[string]string{}

	// When there are no commits, staged files won't appear in `git diff HEAD`,
	// so we capture them here as full file contents.
	if !hasHEAD() {
		out, err := exec.Command("git", "ls-files").Output()
		if err != nil {
			return nil, fmt.Errorf("git ls-files: %w", err)
		}
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f == "" {
				continue
			}
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("reading staged file %s: %w", f, err)
			}
			result[f] = string(data)
		}
	}

	out, err := exec.Command("git", "ls-files", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files --others: %w", err)
	}
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f == "" {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("reading untracked file %s: %w", f, err)
		}
		result[f] = string(data)
	}
	return result, nil
}

func ResetTracked() error {
	if !hasHEAD() {
		// No commits yet. Remove staged files from the working tree first (so
		// git rm --cached won't complain about staged-vs-worktree divergence),
		// then clear the index.
		out, err := exec.Command("git", "ls-files").Output()
		if err != nil {
			return fmt.Errorf("listing staged files: %w", err)
		}
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f != "" {
				os.Remove(f)
			}
		}
		cmd := exec.Command("git", "rm", "-r", "--cached", "--ignore-unmatch", ".")
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command("git", "checkout", "--", ".")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ApplyPatch(patch string) error {
	if patch == "" {
		return nil
	}
	tmp, err := os.CreateTemp("", "checkpoint-*.patch")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(patch); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	cmd := exec.Command("git", "apply", tmp.Name())
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RestoreUntracked removes untracked files from the previous checkpoint that
// don't exist in the target, then writes the target's untracked files.
func RestoreUntracked(prevUntracked, targetUntracked map[string]string) error {
	for path := range prevUntracked {
		if _, ok := targetUntracked[path]; !ok {
			os.Remove(path)
		}
	}
	for path, content := range targetUntracked {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("restoring %s: %w", path, err)
		}
	}
	return nil
}

// CurrentMatchesCheckpoint returns true if the working tree matches cp exactly.
func CurrentMatchesCheckpoint(cp *Checkpoint) (bool, error) {
	diff, err := CaptureDiff()
	if err != nil {
		return false, err
	}
	untracked, err := CaptureUntracked()
	if err != nil {
		return false, err
	}
	if diff != cp.Patch {
		return false, nil
	}
	if len(untracked) != len(cp.Untracked) {
		return false, nil
	}
	for k, v := range untracked {
		if cp.Untracked[k] != v {
			return false, nil
		}
	}
	return true, nil
}
