package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

type HandleResult struct {
	Platform   string `json:"platform"`
	Handle     string `json:"handle"`
	Hijackable bool   `json:"hijackable"`
}

type URLResult struct {
	URL     string         `json:"url"`
	Handles []HandleResult `json:"handles"`
}

var potentialPages = []string{
	"contact", "contact-us", "about", "about-us", "team", "support", "help", "get-in-touch",
	"meet-the-team", "our-team", "company", "who-we-are", "info", "legal", "press", "reach-us",
}

var regexPatterns = map[string]*regexp.Regexp{
	"twitter":   regexp.MustCompile(`https?://(?:www\.)?(?:x\.com|twitter\.com)/([a-zA-Z0-9_\-]+)`),
	"linkedin":  regexp.MustCompile(`https?://(?:www\.)?linkedin\.com/in/([a-zA-Z0-9_\-]+)`),
	"youtube":   regexp.MustCompile(`https?://(?:www\.)?youtube\.com/(?:c|channel|user)/([a-zA-Z0-9_\-]+)`),
	"facebook":  regexp.MustCompile(`https?://(?:www\.)?facebook\.com/([a-zA-Z0-9_\-]+)`),
	"instagram": regexp.MustCompile(`https?://(?:www\.)?instagram\.com/([a-zA-Z0-9_\-]+)`),
	"tiktok":    regexp.MustCompile(`https?://(?:www\.)?tiktok\.com/@([a-zA-Z0-9_\-]+)`),
}

func checkTwitterHandleAvailability(handle string) bool {
	url := "https://x.com/" + handle
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 404 {
			return true
		}
	}
	return false
}

func findHandles(content string) []HandleResult {
	var found []HandleResult
	for platform, regex := range regexPatterns {
		matches := regex.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			handle := m[1]
			if platform == "twitter" {
				found = append(found, HandleResult{Platform: platform, Handle: handle, Hijackable: checkTwitterHandleAvailability(handle)})
			} else {
				found = append(found, HandleResult{Platform: platform, Handle: handle, Hijackable: false})
			}
		}
	}
	return found
}

func fetchAndAnalyze(urlStr string, mutex *sync.Mutex, outputFile *os.File, verbose bool) {
	if verbose {
		fmt.Println("[VERBOSE] Fetching:", urlStr)
	}
	resp, err := http.Get(urlStr)
	if err != nil || resp.StatusCode != 200 {
		if verbose {
			fmt.Printf("[VERBOSE] Failed to fetch main URL: %s, error: %v\n", urlStr, err)
		}
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	content := string(body)

	handles := findHandles(content)

	parsedUrl, _ := url.Parse(urlStr)
	baseUrl := parsedUrl.Scheme + "://" + parsedUrl.Host

	for _, page := range potentialPages {
		pageUrl := baseUrl + "/" + page
		if verbose {
			fmt.Println("[VERBOSE] Fetching potential page:", pageUrl)
		}
		pResp, err := http.Get(pageUrl)
		if err == nil && pResp.StatusCode == 200 {
			pageBody, _ := ioutil.ReadAll(pResp.Body)
			pResp.Body.Close()
			handles = append(handles, findHandles(string(pageBody))...)
		}
	}

	uniqueHandles := make(map[string]HandleResult)
	for _, h := range handles {
		uniqueHandles[h.Platform+":"+h.Handle] = h
	}

	var finalHandles []HandleResult
	for _, h := range uniqueHandles {
		fmt.Printf("[+] %s, @%s (%s), %s\n", urlStr, h.Handle, h.Platform, func() string {
			if h.Hijackable {
				return "HIJACKABLE"
			}
			return "Not Hijackable"
		}())
		finalHandles = append(finalHandles, h)
	}

	partialResult := URLResult{URL: urlStr, Handles: finalHandles}
	mutex.Lock()
	jsonResult, _ := json.MarshalIndent(partialResult, "", "  ")
	outputFile.Write(jsonResult)
	outputFile.Write([]byte(",\n"))
	mutex.Unlock()
}

func main() {
	filePath := flag.String("file", "", "Path to URLs file")
	flag.StringVar(filePath, "f", "", "Path to URLs file (shorthand)")

	threads := flag.Int("threads", 5, "Number of concurrent workers")
	flag.IntVar(threads, "t", 5, "Number of concurrent workers (shorthand)")

	output := flag.String("output", "result.json", "Output JSON file")
	flag.StringVar(output, "o", "result.json", "Output JSON file (shorthand)")

	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.BoolVar(verbose, "v", false, "Enable verbose output (shorthand)")

	flag.Parse()

	if *filePath == "" {
		fmt.Println("Please provide --file or -f argument")
		os.Exit(1)
	}

	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var urls []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	outputFile, err := os.Create(*output)
	if err != nil {
		fmt.Println("Failed to create output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()
	outputFile.Write([]byte("[\n"))

	var wg sync.WaitGroup
	var mutex sync.Mutex
	sem := make(chan struct{}, *threads)

	for i, u := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(u string, idx int) {
			defer func() { <-sem; wg.Done() }()
			fetchAndAnalyze(u, &mutex, outputFile, *verbose)
			if idx == len(urls)-1 {
				mutex.Lock()
				outputFile.Write([]byte("{}\n]")) // Close JSON array
				mutex.Unlock()
			}
		}(u, i)
	}

	wg.Wait()
	fmt.Println("Processing complete. Results written to:", *output)
}

