package profanity

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"encoding/json"
	"io"
	"bytes"
	"sync"
	"log"
)

var wordsMap map[string]interface{}

func init() {
	wordsMap = make(map[string]interface{})
	cacheAbuses()
}

type ProfanityResp struct {
	Total int  `json:"total"`
	Found []string `json:"found"`
}

func Filter(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		channel := make(chan string)
		var wg sync.WaitGroup
		found := []string{}
		text, err := ioutil.ReadAll(req.Body)
		checkErr(err)
		words := strings.Split(string(text), " ")
		wg.Add(len(words))
		go func() {
			for msg := range channel {
				found = append(found, msg)
				wg.Done()
			}
		}()
		for _, word := range words {
			log.Println(word)
			go func(w string) {
				s := strings.TrimSpace(w);
				if _, ok := wordsMap[s]; s != "" && ok {
					channel <- s
				} else {
					wg.Done()
				}
			}(word)
		}
		wg.Wait()
		close(channel)
		profanityResp := ProfanityResp{Total:len(found), Found:found}
		response, err := json.MarshalIndent(profanityResp, "", "    ")
		checkErr(err)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		io.Copy(w, bytes.NewReader(response))
	default:
		http.Error(w, fmt.Sprintf("Unsupported method: %s", req.Method), http.StatusMethodNotAllowed)
	}
}

func Recache(w http.ResponseWriter, req *http.Request) {
	log.Println("Recaching Abuses")
	cacheAbuses()
	log.Println("Reacaching Done")
	fmt.Fprint(w, "Recaching Done")
}

func cacheAbuses() {
	cacheDirContent("data")
}

func cacheDirContent(dir string) {
	files, err := ioutil.ReadDir("data")
	checkErr(err)
	if len(files) > 0 {
		for _, file := range files {
			if file.Mode().IsRegular() && !file.IsDir() {
				f, err := ioutil.ReadFile(dir + "/" + file.Name())
				checkErr(err)
				if err == nil {
					words := strings.Split(string(f), "\n")
					for _, s := range words {
						wordsMap[strings.TrimSpace(s)] = nil
					}
				}
			}
		}
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}
