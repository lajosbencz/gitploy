package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println(os.Args)
	configFile := "config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := ConfigFromYamlFile(configFile)
	if err != nil {
		panic(err)
	}

	hookHandler := &HookHandler{logs: make(chan string), Config: *config}
	go hookHandler.log()

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
