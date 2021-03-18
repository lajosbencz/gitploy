package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	var hookData HookData
	err := json.NewDecoder(r.Body).Decode(&hookData)
	if err != nil {
		handleError(w, err)
		return
	}
	hdFile, err := os.Create("hookdata-" + strings.ReplaceAll(hookData.Project.PathWithNamespace, "/", "-") + "-" + hookData.ObjectKind + "-" + hookData.CheckoutSha + ".json")
	if err != nil {
		handleError(w, err)
		return
	}
	hdFile.Write(bodyBytes)
	log.Println("received [" + hookData.ObjectKind + "] for [" + hookData.Repository.GitHTTPUrl + "]:[" + hookData.GetTag() + "]")

	if hookData.ObjectKind == "push" || hookData.ObjectKind == "tag_push" {

		projectConfig, err := t.Config.GetProjectByRemote(hookData.Repository.GitHTTPUrl)
		if err != nil {
			handleError(w, err)
			return
		}

		if err = HookGitSync(&hookData, projectConfig); err != nil {
			handleError(w, err)
			return
		}

		if err = HookPre(&hookData, projectConfig); err != nil {
			handleError(w, err)
			return
		}

		if err = HookDependencies(&hookData, projectConfig); err != nil {
			handleError(w, err)
			return
		}

		if err = HookBuild(&hookData, projectConfig); err != nil {
			handleError(w, err)
			return
		}

		if err = HookPost(&hookData, projectConfig); err != nil {
			handleError(w, err)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("{\"status\":\"done\"}"))

	} else {
		log.Println("ignored event")
	}
}

func (t *HookHandler) log() {
	for item := range t.logs {
		log.Println(item)
	}
}
