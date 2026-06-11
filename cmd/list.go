package cmd

import (
	"git-checkpoint/internal/checkpoint"
	"git-checkpoint/internal/ui"
)

func runList() error {
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

	isDirty := false
	if meta.Head != "" {
		head, err := checkpoint.LoadCheckpoint(gitDir, meta.Head)
		if err != nil {
			return err
		}
		matches, err := checkpoint.CurrentMatchesCheckpoint(head)
		if err != nil {
			return err
		}
		isDirty = !matches
	}

	selected, err := ui.RunList(chain, meta.Head, isDirty)
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

	if err := applyRestore(gitDir, head, selected); err != nil {
		return err
	}

	if selected.ID != head.ID {
		meta.Head = selected.ID
		if err := checkpoint.SaveMeta(gitDir, meta); err != nil {
			return err
		}
	}

	return nil
}
