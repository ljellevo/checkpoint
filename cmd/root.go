package cmd

import (
	"fmt"
	"time"

	"git-checkpoint/internal/checkpoint"

	"github.com/spf13/cobra"
)

var message string
var listFlag bool

var rootCmd = &cobra.Command{
	Use: "checkpoint",
	Short: "Create a checkpoint of your current working tree state. \n" +
		"This is not a git commit, but a temporary snapshot. \n" +
		"• Use \"checkpoint -l.\" to see and select between the different checkpoints \n" +
		"• Use \"checkpoint restore\" to go back to previous checkpoint directly.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listFlag {
			return runList()
		}

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

		// If a future checkpoint exists branching from the current HEAD, creating
		// a new one here will invalidate it. Warn and require confirmation.
		future, err := checkpoint.FindChild(gitDir, meta.Head)
		if err != nil {
			return err
		}
		if future != nil {
			label := future.Message
			if label == "" {
				label = "(no message)"
			}
			fmt.Printf("Warning: creating this checkpoint will discard [%s] — %s and everything after it.\n", future.ID, label)
			fmt.Print("Continue? [y/N] ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Println("Checkpoint cancelled.")
				return nil
			}
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
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().StringVarP(&message, "message", "m", "", "create checkpoint with message")
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "list all checkpoints")
	rootCmd.AddCommand(restoreCmd)
}
