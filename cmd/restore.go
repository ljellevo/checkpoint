package cmd

import (
	"fmt"

	"git-checkpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore the working tree to the previous checkpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		gitDir, err := checkpoint.GetGitDir()
		if err != nil {
			return err
		}

		meta, err := checkpoint.LoadMeta(gitDir)
		if err != nil {
			return err
		}

		if meta.Head == "" {
			return fmt.Errorf("no checkpoints found")
		}

		head, err := checkpoint.LoadCheckpoint(gitDir, meta.Head)
		if err != nil {
			return err
		}

		// If the working tree doesn't match the HEAD checkpoint, restore TO the
		// HEAD checkpoint (discard uncommitted changes). Only advance to the
		// parent when the working tree already matches HEAD.
		matches, err := checkpoint.CurrentMatchesCheckpoint(head)
		if err != nil {
			return err
		}

		var target *checkpoint.Checkpoint
		if !matches {
			// Dirty — restore to HEAD checkpoint, discarding uncommitted changes.
			target = head
		} else {
			// Clean at HEAD — go one step back to the parent.
			if head.ParentID == "" {
				return fmt.Errorf("already at the oldest checkpoint, nothing to restore")
			}
			target, err = checkpoint.LoadCheckpoint(gitDir, head.ParentID)
			if err != nil {
				return err
			}
		}

		if err := checkpoint.ResetTracked(); err != nil {
			return fmt.Errorf("resetting tracked files: %w", err)
		}

		if err := checkpoint.RestoreUntracked(head.Untracked, target.Untracked); err != nil {
			return err
		}

		if err := checkpoint.ApplyPatch(target.Patch); err != nil {
			return fmt.Errorf("applying patch: %w", err)
		}

		// Only update HEAD when we actually moved to a different checkpoint.
		if target.ID != head.ID {
			meta.Head = target.ID
			if err := checkpoint.SaveMeta(gitDir, meta); err != nil {
				return err
			}
		}

		label := target.Message
		if label == "" {
			label = "(no message)"
		}
		fmt.Printf("✓ Restored to checkpoint [%s] — %s\n", target.ID, label)
		return nil
	},
}
