package main

import (
	"os"
	"testing"
)

const localPath = "/var/www/gitploy-dummy"
const remotePath = "https://github.com/lajosbencz/dummy"

func TestGitOps(t *testing.T) {
	err := os.RemoveAll(localPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gitClone(remotePath, localPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	if _, err = os.Stat(localPath + "/README.md"); os.IsNotExist(err) {
		t.Fatal("README.md not found in repo")
	}
	err = gitPull(localPath)
	if err != nil {
		t.Fatal(err.Error())
	}
	if _, err = os.Stat(localPath + "/README.md"); os.IsNotExist(err) {
		t.Fatal("README.md not found in repo")
	}
	err = os.RemoveAll(localPath)
	if err != nil {
		t.Fatal(err.Error())
	}
}
