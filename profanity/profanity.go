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
	"regexp"
)

var blank *regexp.Regexp
var uRepeat *regexp.Regexp
var iRepeat *regexp.Regexp

var wordsMap map[string]interface{}

func init() {
	wordsMap = make(map[string]interface{})
	cacheAbuses()
	regexpCompile()
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
		b, err := ioutil.ReadAll(req.Body)
		text := filterUsingRegex(string(b))
		checkErr(err)
		log.Println(text)
		words := strings.Split(text, " ")
		wg.Add(len(words))
		go func() {
			for msg := range channel {
				found = append(found, msg)
				wg.Done()
			}
		}()
		for _, word := range words {
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

func Heartbeat(w http.ResponseWriter, req * http.Request) {}

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

func regexpCompile() {
	blank = regexp.MustCompile("([\\[$&,:;=?#|'<>.^*\\(\\)%\\]])|(\\b\\d+\\b)|(cum\u0020laude)|(he\\'ll)|(\\B\\#)|(&\\#?[a-z0-9]{2,8};)|(\\b\\'+)|(\\'+\\b)|(\\b\\\")|(\\\"\\b)|(dick\u0020cheney)|(\\!+\\B)")
	uRepeat = regexp.MustCompile("u+")
	iRepeat = regexp.MustCompile("i+")
}

func filterUsingRegex(text string) string {
	text = blank.ReplaceAllString(text, "")
	text = uRepeat.ReplaceAllString(text, "u")
	text = iRepeat.ReplaceAllString(text, "i")
	return text
}