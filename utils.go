package main

import "os"

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
