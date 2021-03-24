package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type npmPackageScripts struct {
	Scripts map[string]string `json:"scripts"`
}

type HookHandler struct {
	Config Config
}

func handleError(w http.ResponseWriter, err error, prefix string) {
	log.Println(prefix, err.Error())
	http.Error(w, strings.TrimLeft(prefix+" "+err.Error(), " "), 500)
}

func hookActors(hookData HookData, projectConfig ConfigProject) error {
	if err := HookGitSync(hookData, projectConfig); err != nil {
		log.Println("error at git sync hook:", err.Error())
		return err
	}
	if err := HookPre(hookData, projectConfig); err != nil {
		log.Println("error at pre hook:", err.Error())
		return err
	}
	if err := HookDependencies(hookData, projectConfig); err != nil {
		log.Println("error at dependencies hook:", err.Error())
		return err
	}
	if err := HookBuild(hookData, projectConfig); err != nil {
		log.Println("error at build hook:", err.Error())
		return err
	}
	if err := HookPost(hookData, projectConfig); err != nil {
		log.Println("error at post hook:", err.Error())
		return err
	}
	return nil
}

func (t *HookHandler) handleHook(w http.ResponseWriter, r *http.Request) {
	whitelist := strings.Split("127.0.0.1,::1", ",")
	remoteAddr := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		remoteAddr = forwardedFor
	}
	if r.Header.Get(t.Config.Token.Key) != t.Config.Token.Value && !stringListContains(whitelist, remoteAddr) {
		handleError(w, fmt.Errorf("failed to verify token"), "")
		return
	}
	log.Println("HookHandler: " + r.RequestURI)
	var hookData HookData
	err := json.NewDecoder(r.Body).Decode(&hookData)
	if err != nil {
		handleError(w, err, "failed to parse hook data")
		return
	}
	log.Println("received [" + hookData.ObjectKind + "] for [" + hookData.Repository.GitHTTPUrl + "]:[" + hookData.GetTag() + "]")

	constraintPassed := true

	projectConfig, err := t.Config.GetProjectByRemote(hookData.Repository.GitHTTPUrl)
	if err != nil {
		log.Println("ignoring project:", hookData.Repository.GitHTTPUrl)
		constraintPassed = false
	} else if hookData.ObjectKind != "push" && hookData.ObjectKind != "tag_push" {
		log.Println("ignoring hook object kind:", hookData.ObjectKind)
		constraintPassed = false
	} else {
		branchOrTag := hookData.GetTag()
		if projectConfig.Mode == "branch" {
			if projectConfig.Constraint != branchOrTag {
				log.Println("ignoring branch:", branchOrTag)
				constraintPassed = false
			}
		} else if projectConfig.Mode == "semver" {
			svCon, err := semver.NewConstraint(projectConfig.Constraint)
			if err != nil {
				handleError(w, err, "failed to create semver constraint")
				return
			}
			svVer, err := semver.NewVersion(branchOrTag)
			if err != nil {
				handleError(w, err, "failed to create semver version")
				return
			}
			if !svCon.Check(svVer) {
				log.Println("ignoring semver:", branchOrTag)
				constraintPassed = false
			}
		}
	}

	w.Header().Add("Content-Type", "application/json")
	if constraintPassed {
		go hookActors(hookData, *projectConfig)
		w.Write([]byte("{\"status\":\"ok\"}"))
		log.Println("done")
		return
	}
	w.Write([]byte("{\"status\":\"ignored\"}"))
	log.Println("ignored")
}
