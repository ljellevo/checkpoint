package cmd

import (
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

		chain, err := checkpoint.WalkChain(gitDir, meta.Head)
		if err != nil {
			return err
		}

		return ui.RunList(chain)
	},
}
