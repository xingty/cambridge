package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Index struct {
	start int32
	end   int32
}

func (idx Index) Stringer() string {
	return fmt.Sprintf("start: %d, end: %d\n", idx.start, idx.end)
}

var indexing map[string]Index
var dictFile *os.File

func readIndex(dictBasePath string) []string {
	binPath := filepath.Join(dictBasePath, "index.bin")
	data, err := os.ReadFile(binPath)
	if err != nil {
		panic(err)
	}

	return strings.Split(string(data), "\n")
}

func loadIndex(dictBasePath string) {
	indexing = make(map[string]Index)
	lines := readIndex(dictBasePath)
	for _, line := range lines {
		segs := strings.Split(line, "|")
		if len(segs) != 3 {
			continue
		}

		num, err := strconv.Atoi(segs[1])
		if err != nil {
			fmt.Printf("wrong line: %s\n", line)
			continue
		}
		start := num
		num, err = strconv.Atoi(segs[2])
		if err != nil {
			fmt.Printf("wrong line: %s\n", line)
			continue
		}
		end := num

		idx := Index{
			start: int32(start),
			end:   int32(end),
		}

		indexing[segs[0]] = idx
	}
}

func openDict(dictBasePath string) {
	dictPath := filepath.Join(dictBasePath, "dict.bin")
	var err error
	dictFile, err = os.Open(dictPath)
	if err != nil {
		panic(err)
	}
}

func query(start, end int32) (string, error) {
	length := end - start + 1
	buffer := make([]byte, length)

	n, err := dictFile.ReadAt(buffer, int64(start))
	if n != int(length) || err != nil {
		return "", fmt.Errorf("failed to read from file: %v", err)
	}

	return string(buffer), nil
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "Keyword is required", http.StatusBadRequest)
		return
	}

	idx, exists := indexing[keyword]
	if !exists {
		http.Error(w, "Word not found", http.StatusNotFound)
		return
	}

	content, err := query(idx.start, idx.end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json;chatset=utf-8")
	w.Write([]byte(content + "\n"))
}

func startServer(addr string) {
	http.HandleFunc("/search", handleSearch)
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func main() {
	var port int
	var ipAddr string
	var dictBasePath string

	flag.IntVar(&port, "port", 8010, "port number")
	flag.StringVar(&ipAddr, "addr", "127.0.0.1", "ip address")
	flag.StringVar(&dictBasePath, "dict", "", "dict base path")

	flag.Parse()

	loadIndex(dictBasePath)
	openDict(dictBasePath)
	addr := fmt.Sprintf("%s:%d", ipAddr, port)
	startServer(addr)

	defer dictFile.Close()
}
