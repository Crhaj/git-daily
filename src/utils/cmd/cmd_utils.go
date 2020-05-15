package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func Pwd() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get path to working directory", err)
	}
	return filepath.FromSlash(path)
}

func GetDirContent(path string) []os.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal("Failed on command GetDirContent", err)
	}
	fmt.Println("Files and directories in current folder:")
	for _, file := range files {
		fmt.Printf("name: %v isDir: %v\n", file.Name(), file.IsDir())
	}
	return files
}

func GetDirectories(files []os.FileInfo) []os.FileInfo {
	var dirs []os.FileInfo
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file)
		}
	}
	return dirs
}
