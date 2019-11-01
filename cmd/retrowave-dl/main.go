package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	limitFlag = flag.Int("limit", 2, "tracks number for download")
	allFlag   = flag.Bool("all", false, "get all possible tracks (ignoring --limit flag)")
	jsonFlag  = flag.Bool("json", false, "download track list as JSON file")
	syncFlag  = flag.Bool("sync", false, "synchronize downloaded files")
	outFlag   = flag.String("out", "", "directory for output")

	downloadDir string
)

const (
	baseUrl = "http://retrowave.ru"
)

type Response struct {
	Status int `json:"status"`
	Body   struct {
		Cursor int     `json:"cursor"`
		Tracks []Track `json:"tracks"`
	} `json:"body"`
}

type Track struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Duration   float64 `json:"duration"`
	StreamURL  string  `json:"streamUrl"`
	ArtworkURL string  `json:"artworkUrl"`
}

type JsonOutput []map[string]string

func getTracks(limit int, out chan *Response) {
	url := fmt.Sprintf(baseUrl+"/api/v1/tracks?cursor=1&limit=%d", limit)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Host", "retrowave.ru")
	req.Header.Add("Referer", "http://retrowave.ru/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")

	res, err := http.DefaultClient.Do(req)
	if res != nil {
		defer func() {
			if errr := res.Body.Close(); errr != nil {
				log.Fatal(errr)
			}
		}()
	}
	if err != nil {
		log.Fatal(err)
	}
	body, _ := ioutil.ReadAll(res.Body)

	resp := &Response{}
	if err := json.Unmarshal(body, resp); err != nil {
		log.Fatal(err)
	}
	out <- resp
}

func createJson(r *Response) {
	output := make(JsonOutput, 0)
	for _, track := range r.Body.Tracks {
		if track.ID == "" {
			continue
		}
		output = append(output, map[string]string{
			"id":       track.ID,
			"title":    track.Title,
			"link":     baseUrl + track.StreamURL,
			"filename": track.Title + ".mp3",
		})
	}
	if b, err := json.MarshalIndent(output, "", "  "); err == nil {
		if _, errDir := os.Stat(downloadDir); errDir != nil {
			if errDirMake := os.MkdirAll(downloadDir, os.ModePerm); errDirMake != nil {
				log.Fatal(errDirMake)
			}
		}
		path := filepath.Join(downloadDir, "soundtracks.json")
		err = ioutil.WriteFile(path, b, 0644)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("JSON file saved at '%s'\n", path)
	}

}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(f string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if resp != nil {
		defer func() {
			if errr := resp.Body.Close(); errr != nil {
				log.Fatal(errr)
			}
		}()
	}
	if err != nil {
		return err
	}

	// Create the file
	f = filepath.Join(downloadDir, f)
	out, err := os.Create(f)
	if out != nil {
		defer func() {
			if errf := out.Close(); errf != nil {
				log.Fatal(errf)
			}
		}()
	}
	if err != nil {
		return err
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func downloadFiles(r *Response) {
	var wg sync.WaitGroup
	var skipNum uint64
	var downloadNum uint64
	for _, track := range r.Body.Tracks {
		if track.ID == "" {
			continue
		}
		name := track.Title + ".mp3"
		// strip slashes
		name = strings.Replace(name, "/", "\\", -1)

		if *syncFlag == true {
			if _, err := os.Stat(filepath.Join(downloadDir, name)); err == nil {
				log.Printf("[SKIP] %s\n", name)
				atomic.AddUint64(&skipNum, 1)
				continue
			}
		}

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			if err := DownloadFile(name, baseUrl+track.StreamURL); err != nil {
				log.Fatal(err)
			}
			atomic.AddUint64(&downloadNum, 1)
			log.Printf("[DOWNLOAD] %s\n", name)
			wg.Done()
		}(&wg)
		time.Sleep(time.Second)
	}
	wg.Wait()

	downloadNumFinal := atomic.LoadUint64(&downloadNum)
	skipNumFinal := atomic.LoadUint64(&skipNum)
	log.Printf("Downloaded: %d, Skiped: %d\n", downloadNumFinal, skipNumFinal)
}

func main() {
	flag.Parse()
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if *outFlag == "" {
		downloadDir = filepath.Join(dir, "downloads")
	} else {
		downloadDir = *outFlag
	}

	limit := *limitFlag
	if *allFlag == true {
		limit = 999
	}
	doneResp := make(chan *Response, 1)
	go getTracks(limit, doneResp)

	resp := <-doneResp

	if *jsonFlag == true {
		createJson(resp)
		os.Exit(1)
	}

	downloadFiles(resp)
}
