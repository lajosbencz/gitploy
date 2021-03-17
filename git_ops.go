package main

import (
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

func gitSync(remote string, local string) error {
	if _, osErr := os.Stat(local); os.IsNotExist(osErr) {
		if err := gitClone(remote, local); err != nil {
			return err
		}
	} else {
		if err := gitPull(local); err != nil {
			return err
		}
	}
	return nil
}

func gitClone(remote string, local string) error {
	_, err := git.PlainClone(local, false, &git.CloneOptions{
		URL:      remote,
		Depth:    1,
		Progress: os.Stdout,
	})
	return err
}

func gitPull(local string) error {
	r, err := git.PlainOpen(local)
	if err != nil {
		return err
	}
	err = r.Fetch(&git.FetchOptions{Progress: os.Stdout})
	if err != nil && !strings.Contains(err.Error(), "already up-to-date") {
		return err
	}
	h, err := r.Head()
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	err = w.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: h.Hash()})
	if err != nil {
		return err
	}
	return nil
}
