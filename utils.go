package main

import (
	"os"
	"os/exec"
)

func isDir(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	mode := stat.Mode()
	return mode.IsDir()
}

func isFile(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	mode := stat.Mode()
	return mode.IsRegular()
}

func runCmd(bin string, args ...string) error {
	exec := exec.Command(bin, args...)
	exec.Stdout = os.Stdout
	exec.Stderr = os.Stderr
	err := exec.Run()
	if err != nil {
		return err
	}
	return nil
}

func stringListContains(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}
