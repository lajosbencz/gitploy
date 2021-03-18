package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	configFile := "gitploy.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := ConfigFromYamlFile(configFile)
	if err != nil {
		panic(err)
	}

	hookHandler := &HookHandler{Config: *config}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		hookHandler.handleHook(w, r)
	})

	fmt.Println("listening on " + config.Listen)
	http.ListenAndServe(config.Listen, nil)
}
