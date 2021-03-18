package main

import (
	"encoding/json"
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

func (t *HookHandler) handleHook(w http.ResponseWriter, r *http.Request) {
	log.Println("HookHandler: " + r.RequestURI)
	// bodyBytes, _ := ioutil.ReadAll(r.Body)
	// r.Body.Close()
	// r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	var hookData HookData
	err := json.NewDecoder(r.Body).Decode(&hookData)
	if err != nil {
		handleError(w, err, "failed to parse hook data")
		return
	}
	// hdFile, err := os.Create("hookdata-" + strings.ReplaceAll(hookData.Project.PathWithNamespace, "/", "-") + "-" + hookData.ObjectKind + "-" + hookData.CheckoutSha + ".json")
	// if err != nil {
	// 	handleError(w, err)
	// 	return
	// }
	// hdFile.Write(bodyBytes)
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
		log.Println("constraints passed")
		if err = HookGitSync(&hookData, projectConfig); err != nil {
			handleError(w, err, "failed to sync git")
			return
		}
		if err = HookPre(&hookData, projectConfig); err != nil {
			handleError(w, err, "failed to run pre hooks")
			return
		}
		if err = HookDependencies(&hookData, projectConfig); err != nil {
			handleError(w, err, "failed to install dependencies")
			return
		}
		if err = HookBuild(&hookData, projectConfig); err != nil {
			handleError(w, err, "failed to build project")
			return
		}
		if err = HookPost(&hookData, projectConfig); err != nil {
			handleError(w, err, "failed to run post hooks")
			return
		}
		w.Write([]byte("{\"status\":\"done\"}"))
		log.Println("done")
		return
	}
	w.Write([]byte("{\"status\":\"ignored\"}"))
	log.Println("ignored")
}
