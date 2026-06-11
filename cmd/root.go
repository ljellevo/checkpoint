package cmd

import (
	"fmt"
	"time"

	"git-checkpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

var message string

var rootCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Create a checkpoint of your current working tree state",
	RunE: func(cmd *cobra.Command, args []string) error {
		gitDir, err := checkpoint.GetGitDir()
		if err != nil {
			return err
		}

		patch, err := checkpoint.CaptureDiff()
		if err != nil {
			return err
		}

		untracked, err := checkpoint.CaptureUntracked()
		if err != nil {
			return err
		}

		meta, err := checkpoint.LoadMeta(gitDir)
		if err != nil {
			return err
		}

		id, err := checkpoint.GenerateID()
		if err != nil {
			return err
		}

		cp := &checkpoint.Checkpoint{
			ID:        id,
			ParentID:  meta.Head,
			Message:   message,
			Timestamp: time.Now().UTC(),
			Patch:     patch,
			Untracked: untracked,
		}

		// Prune any checkpoints that are no longer in the chain from the current
		// head before branching off with a new one.
		if err := checkpoint.PruneOrphans(gitDir, meta.Head); err != nil {
			return err
		}

		if err := checkpoint.SaveCheckpoint(gitDir, cp); err != nil {
			return err
		}

		meta.Head = cp.ID
		if err := checkpoint.SaveMeta(gitDir, meta); err != nil {
			return err
		}

		ts := cp.Timestamp.Local().Format("2006-01-02 15:04:05")
		if cp.Message != "" {
			fmt.Printf("✓ Checkpoint [%s] created at %s — %s\n", cp.ID, ts, cp.Message)
		} else {
			fmt.Printf("✓ Checkpoint [%s] created at %s\n", cp.ID, ts)
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&message, "message", "m", "", "Checkpoint message")
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
}
