package cmd

import (
	"fmt"

	"git-checkpoint/internal/checkpoint"
	"git-checkpoint/internal/ui"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all checkpoints",
	RunE: func(cmd *cobra.Command, args []string) error {
		gitDir, err := checkpoint.GetGitDir()
		if err != nil {
			return err
		}

		meta, err := checkpoint.LoadMeta(gitDir)
		if err != nil {
			return err
		}

		chain, err := checkpoint.BuildFullChain(gitDir, meta.Head)
		if err != nil {
			return err
		}

		selected, err := ui.RunList(chain, meta.Head)
		if err != nil {
			return err
		}
		if selected == nil {
			return nil
		}

		head, err := checkpoint.LoadCheckpoint(gitDir, meta.Head)
		if err != nil {
			return err
		}

		// Prompt if the working tree has unsaved changes.
		ok, err := confirmIfDirty(head)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Restore cancelled.")
			return nil
		}

		if err := applyRestore(gitDir, head, selected); err != nil {
			return err
		}

		if selected.ID != head.ID {
			meta.Head = selected.ID
			if err := checkpoint.SaveMeta(gitDir, meta); err != nil {
				return err
			}
		}

		label := selected.Message
		if label == "" {
			label = "(no message)"
		}
		fmt.Printf("✓ Restored to checkpoint [%s] — %s\n", selected.ID, label)
		return nil
	},
}
