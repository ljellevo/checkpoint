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

		matches, err := checkpoint.CurrentMatchesCheckpoint(head)
		if err != nil {
			return err
		}

		var target *checkpoint.Checkpoint

		if !matches {
			// Dirty — prompt, then restore to HEAD checkpoint.
			ok, err := confirmIfDirty(head)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("Restore cancelled.")
				return nil
			}
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

		if err := applyRestore(gitDir, head, target); err != nil {
			return err
		}

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
