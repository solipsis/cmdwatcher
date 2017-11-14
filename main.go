package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type cache struct {
	sync.RWMutex
	internal map[string]time.Time
}

type fileInfo struct {
	path string
	ts   time.Time
}

type op int32

const (
	Edit op = iota
	Add
	Remove
)

// TODO: implement add and remove
var ops = map[op]string{
	Edit:   "Edit",
	Add:    "Add",
	Remove: "Remove",
}

func (o op) String() string {
	if str, ok := ops[o]; ok {
		return str
	}
	return "Uknown Op"
}

func newCache() *cache {
	return &cache{
		internal: make(map[string]time.Time),
	}
}

func main() {

	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	info := make(chan fileInfo)
	start := time.Now()
	defer func() {
		fmt.Printf("elapsed time: %v\n", time.Since(start))
	}()
	scanRoots(roots, info)
	cache := newCache()
	for file := range info {
		cache.Lock()
		cache.internal[file.path] = file.ts
		cache.Unlock()
	}
	fmt.Println("Existing files cached. Watching for changes...")
	poll(cache, roots)
}

// scan directories starting at entries in roots
func scanRoots(roots []string, res chan<- fileInfo) {
	var wg sync.WaitGroup
	for _, root := range roots {
		wg.Add(1)
		go walkDir(root, &wg, res)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

}

// Check for changes every interval. TODO: make command line flag
func poll(c *cache, roots []string) {
	tick := time.Tick(1000 * time.Millisecond)
	for range tick {
		//	fmt.Println("tick")
		if detectChanges(c, roots) {
			gulp()
		}
	}
}

func detectChanges(c *cache, roots []string) (change bool) {
	res := make(chan fileInfo)
	scanRoots(roots, res)
	for info := range res {
		if c.internal[info.path] != info.ts {
			fmt.Println("file change detected", info.path)
			change = true
			c.Lock()
			c.internal[info.path] = info.ts
			c.Unlock()
		}
	}

	return change
}

// recursively walk directories
func walkDir(dir string, wg *sync.WaitGroup, info chan<- fileInfo) {
	defer wg.Done()
	for _, entry := range dirents(dir) {
		if strings.HasSuffix(entry.Name(), ".swp") {
			continue
		} // TODO make ignore behavior more robust
		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go walkDir(subdir, wg, info)
		} else {
			info <- fileInfo{filepath.Join(dir, entry.Name()), entry.ModTime()}
		}
	}
}

// sema is a counting semaphore to limit concurrency in dirents
// so we don't run out of file headers
var sema = make(chan struct{}, 20)

// TODO make command passable from command line
func gulp() {
	cmd := exec.Command("gulp", "build")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("error during gulp:%v\n", err)
	}
}

// get entries within a directory
func dirents(dir string) []os.FileInfo {
	sema <- struct{}{}
	defer func() { <-sema }()

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdwatcher: %v\n", err)
		log.Fatal(err)
	}
	return entries
}
