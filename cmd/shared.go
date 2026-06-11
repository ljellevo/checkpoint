package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"git-checkpoint/internal/checkpoint"
)

// confirmIfDirty returns true if it's safe to proceed with a restore.
// When the working tree has unsaved changes, the user is asked to confirm.
func confirmIfDirty(head *checkpoint.Checkpoint) (bool, error) {
	matches, err := checkpoint.CurrentMatchesCheckpoint(head)
	if err != nil {
		return false, err
	}
	if matches {
		return true, nil
	}
	fmt.Println("Warning: you have changes not saved in any checkpoint.")
	fmt.Print("They will be discarded. Continue? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(answer)) == "y", nil
}

// applyRestore resets the working tree and applies the target checkpoint.
func applyRestore(gitDir string, from, target *checkpoint.Checkpoint) error {
	if err := checkpoint.ResetTracked(); err != nil {
		return fmt.Errorf("resetting tracked files: %w", err)
	}
	if err := checkpoint.RestoreUntracked(from.Untracked, target.Untracked); err != nil {
		return err
	}
	if err := checkpoint.ApplyPatch(target.Patch); err != nil {
		return fmt.Errorf("applying patch: %w", err)
	}
	return nil
}
