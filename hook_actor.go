package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type HookActor func(HookData, ConfigProject) error

var HookGitSync HookActor = func(hd HookData, cp ConfigProject) error {
	if !isDir(cp.Local) {
		return fmt.Errorf("project must be initialized manually first")
	}
	os.Chdir(cp.Local)
	err := runCmd("git", "fetch", "--all")
	if err != nil {
		return err
	}
	err = runCmd("git", "reset", "--hard", "origin/"+hd.GetTag())
	if err != nil {
		return err
	}
	return nil
}

func execCommand(command []string) error {
	cmdLen := len(command)
	if cmdLen < 1 {
		return fmt.Errorf("empty command")
	}
	log.Println("exec", command)
	if cmdLen < 2 {
		err := runCmd("bash", "-c", command[0])
		if err != nil {
			return err
		}
	}
	return runCmd(command[0], command[1:]...)
}

func execPrePost(localPath string, commandList [][]string) error {
	os.Chdir(localPath)
	for _, cmd := range commandList {
		err := execCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

var HookPre HookActor = func(hd HookData, cp ConfigProject) error {
	os.Chdir(cp.Local)
	return execPrePost(cp.Local, cp.Pre)
}

var HookDependencies HookActor = func(hd HookData, cp ConfigProject) error {
	os.Chdir(cp.Local)
	var wg sync.WaitGroup
	var errComposer, errNpm error
	wg.Add(2)
	go func() {
		if *cp.Integrate.Composer {
			errComposer = runCmd("composer", "install", "--no-interaction")
		}
		wg.Done()
	}()
	go func() {
		if *cp.Integrate.Npm {
			errNpm = runCmd("npm", "install")
		}
		wg.Done()
	}()

	wg.Wait()
	if errComposer != nil {
		return errComposer
	}
	if errNpm != nil {
		return errNpm
	}
	return nil
}

var HookBuild HookActor = func(hd HookData, cp ConfigProject) error {
	os.Chdir(cp.Local)
	if *cp.Integrate.Npm {
		if _, osErr := os.Stat("package.json"); osErr == nil {
			pr, err := os.Open("package.json")
			if err != nil {
				return err
			}
			var npmPackage npmPackageScripts
			err = json.NewDecoder(pr).Decode(&npmPackage)
			if err != nil {
				return err
			}
			if _, ok := npmPackage.Scripts[cp.Integrate.NpmScriptKey]; ok {
				err = runCmd("npm", "run", cp.Integrate.NpmScriptKey)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var HookPost HookActor = func(hd HookData, cp ConfigProject) error {
	os.Chdir(cp.Local)
	return execPrePost(cp.Local, cp.Post)
}
