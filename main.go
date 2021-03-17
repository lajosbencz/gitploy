package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func handleInit(w http.ResponseWriter, r *http.Request) {
	fmt.Println("InitHandler: " + r.RequestURI)
	query := r.URL.Query()
	remote := query.Get("remote")
	local := query.Get("local")
	err := os.RemoveAll(local)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = gitClone(remote, local)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(fmt.Sprintf("remote:\t%s\nlocal:\t%s", remote, local)))
}

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
	http.HandleFunc("/init", handleInit)
	fmt.Println("listening on " + config.Listen)
	http.ListenAndServe(config.Listen, nil)
}

// func longRunningTask(sleep time.Duration) <-chan int32 {
// 	r := make(chan int32)

// 	go func() {
// 		defer close(r)

// 		// Simulate a workload.
// 		time.Sleep(sleep)
// 		r <- rand.Int31n(100)
// 	}()

// 	return r
// }
