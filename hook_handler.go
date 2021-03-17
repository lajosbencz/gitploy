package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
)

type npmPackageScripts struct {
	Scripts map[string]string `json:"scripts"`
}

type HookHandler struct {
	logs   chan string
	Config Config
}

func handleError(w http.ResponseWriter, err error) {
	log.Println(err.Error())
	http.Error(w, err.Error(), 500)
}

func (t *HookHandler) handleHook(w http.ResponseWriter, r *http.Request) {
	log.Println("HookHandler: " + r.RequestURI)
	var hookData HookData
	err := json.NewDecoder(r.Body).Decode(&hookData)
	if err != nil {
		handleError(w, err)
		return
	}
	log.Println("received [" + hookData.ObjectKind + "] for [" + hookData.Repository.GitHTTPUrl + "]:[" + hookData.GetTag() + "]")

	if hookData.ObjectKind == "push" || hookData.ObjectKind == "tag_push" {
		projectConfig, err := t.Config.GetProjectByRemote(hookData.Repository.GitHTTPUrl)
		if err != nil {
			handleError(w, err)
			return
		}
		err = gitSync(projectConfig.Remote, projectConfig.Local)
		if err != nil {
			handleError(w, err)
			return
		}
		log.Println("git synced, changed directory to " + projectConfig.Local)
		os.Chdir(projectConfig.Local)

		for _, cmd := range projectConfig.Pre {
			if len(cmd) < 2 {
				continue
			}
			scr := cmd[1]
			if _, osErr := os.Stat(scr); os.IsNotExist(osErr) {
				continue
			}
			log.Println("exec pre ", cmd)
			exe := exec.Command(cmd[0], cmd[1:]...)
			err = exe.Run()
			if err != nil {
				handleError(w, err)
				return
			}
		}

		var wg sync.WaitGroup
		var errComposer, errNpm error
		wg.Add(2)
		go func() {
			if *projectConfig.Integrate.Composer && (hookData.HasFileChanged("composer.json") || hookData.HasFileChanged("composer.lock")) {
				cmd := exec.Command("composer", "install", "--no-interaction")
				log.Println("install with composer")
				errComposer = cmd.Run()
			}
			wg.Done()
		}()
		go func() {
			if *projectConfig.Integrate.Npm && (hookData.HasFileChanged("package.json") || hookData.HasFileChanged("package-lock.json")) {
				cmd := exec.Command("npm", "install")
				log.Println("install with npm")
				errNpm = cmd.Run()
			}
			wg.Done()
		}()

		wg.Wait()
		if errComposer != nil {
			handleError(w, err)
			return
		}
		if errNpm != nil {
			handleError(w, err)
			return
		}

		if *projectConfig.Integrate.Npm {
			if _, osErr := os.Stat("package.json"); osErr == nil {
				pr, err := os.Open("package.json")
				if err != nil {
					handleError(w, err)
					return
				}
				var npmPackage npmPackageScripts
				err = json.NewDecoder(pr).Decode(&npmPackage)
				if err != nil {
					handleError(w, err)
					return
				}
				if _, ok := npmPackage.Scripts[projectConfig.Integrate.NpmScriptKey]; ok {
					log.Println("executing npm script")
					exe := exec.Command("npm", "run", projectConfig.Integrate.NpmScriptKey)
					err = exe.Run()
					if err != nil {
						handleError(w, err)
						return
					}
				}
			}
		}

		for _, cmd := range projectConfig.Post {
			if len(cmd) < 2 {
				continue
			}
			scr := cmd[1]
			if _, osErr := os.Stat(scr); os.IsNotExist(osErr) {
				continue
			}
			log.Println("exec post ", cmd)
			exe := exec.Command(cmd[0], cmd[1:]...)
			err = exe.Run()
			if err != nil {
				handleError(w, err)
				return
			}

		}
	} else {
		log.Println("ignored event")
	}
}

func (t *HookHandler) log() {
	for item := range t.logs {
		log.Println(item)
	}
}
