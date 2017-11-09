package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	filenames := make(chan string)
	start := time.Now()
	defer func() {
		fmt.Printf("elapsed time: %v\n", time.Since(start))
	}()
	var wg sync.WaitGroup
	for _, root := range roots {
		wg.Add(1)
		go walkDir(root, &wg, filenames)
	}
	go func() {
		wg.Wait()
		close(filenames)
	}()
	/*
		for name := range filenames {
			//	fmt.Println(name)
		}
	*/
	fmt.Println("vim-go")
}

func walkDir(dir string, wg *sync.WaitGroup, filenames chan<- string) {
	defer wg.Done()
	for _, entry := range dirents(dir) {
		filenames <- entry.Name()
		if entry.IsDir() {
			wg.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			go walkDir(subdir, wg, filenames)
		}
	}
}

func dirents(dir string) []os.FileInfo {

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cmdwatcher: %v\n", err)
		log.Fatal(err)
	}
	return entries
}
