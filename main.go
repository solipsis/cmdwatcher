package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()
	dirs := flag.Args()
	fmt.Println("dirs", dirs)
	//infos, err := ioutil.ReadDir(dirs[0])
	walkDir(dirs[0])
	fmt.Println("vim-go")
}

func walkDir(dir string) {
	for _, entry := range dirents(dir) {
		fmt.Println(entry.Name())
		if entry.IsDir() {
			subdir := filepath.Join(dir, entry.Name())
			walkDir(subdir)
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
