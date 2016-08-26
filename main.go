package main

import (
	"net/http"
	"os"
	"log"
	"strconv"
	"github.com/vikramjakhr/purifier/profanity"
	"runtime"
)

const (
	defaultPort string = "9000"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.HandleFunc("/", profanity.Filter)
	http.HandleFunc("/recache", profanity.Recache)
	http.HandleFunc("/heartbeat", profanity.Heartbeat)
	port := port(os.Args)
	log.Println("Server started: http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

func port(args []string) string {
	port := defaultPort
	if len(args) > 1 && args[1] != "" {
		port = args[1]
		_, err := strconv.ParseUint(port, 10, 0)
		if err != nil {
			port = defaultPort
			log.Println("Incorrect port number. Falling back to default : ", defaultPort)
		}
	}
	return port
}