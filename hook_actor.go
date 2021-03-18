package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

type HookActor func(*HookData, *ConfigProject) error

var HookGitSync HookActor = func(hd *HookData, cp *ConfigProject) error {
	return gitSync(cp.Remote, cp.Local)
}

func execCommand(command []string) error {
	cmdLen := len(command)
	if cmdLen < 1 {
		return fmt.Errorf("empty command")
	}
	if cmdLen < 2 {
		exec := exec.Command("bash", "-c", command[0])
		err := exec.Run()
		if err != nil {
			return err
		}
	}
	scr := command[1]
	if _, osErr := os.Stat(scr); os.IsNotExist(osErr) {
		return osErr
	}
	log.Println("exec pre ", command)
	exe := exec.Command(command[0], command[1:]...)
	return exe.Run()
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

var HookPre HookActor = func(hd *HookData, cp *ConfigProject) error {
	os.Chdir(cp.Local)
	return execPrePost(cp.Local, cp.Pre)
}

var HookDependencies HookActor = func(hd *HookData, cp *ConfigProject) error {
	os.Chdir(cp.Local)
	var wg sync.WaitGroup
	var errComposer, errNpm error
	wg.Add(2)
	go func() {
		if *cp.Integrate.Composer {
			var cmd *exec.Cmd
			if hd.HasFileChanged("composer.lock") {
				cmd = exec.Command("composer", "install", "--no-interaction")
			} else if hd.HasFileChanged("composer.json") {
				cmd = exec.Command("composer", "update", "--no-interaction")
			}
			if cmd != nil {
				log.Println("install with composer")
				errComposer = cmd.Run()
			}
		}
		wg.Done()
	}()
	go func() {
		if *cp.Integrate.Npm && (hd.HasFileChanged("package.json") || hd.HasFileChanged("package-lock.json")) {
			cmd := exec.Command("npm", "install")
			log.Println("install with npm")
			errNpm = cmd.Run()
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

var HookBuild HookActor = func(hd *HookData, cp *ConfigProject) error {
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
				log.Println("executing npm script")
				exe := exec.Command("npm", "run", cp.Integrate.NpmScriptKey)
				err = exe.Run()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

var HookPost HookActor = func(hd *HookData, cp *ConfigProject) error {
	os.Chdir(cp.Local)
	return execPrePost(cp.Local, cp.Post)
}
