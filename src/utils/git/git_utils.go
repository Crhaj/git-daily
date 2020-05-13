package git

import (
	"fmt"
	"git-daily/src/utils/common"
	"log"
	"os"
	"os/exec"
)

type IsGitRepoResult struct {
	IsGitRepo bool
	path      string
}

func IsDirGitRepo(path string, ch chan<- IsGitRepoResult) {
	out, err := exec.Command("git",  "-C", path, "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil {
		ch <- IsGitRepoResult{IsGitRepo: false, path: path}
	}
	isRepo, parseErr := common.ParseBoolFromBytes(out)
	if parseErr != nil {
		fmt.Println("Error parsing check result for:", path, parseErr)
		ch <- IsGitRepoResult{IsGitRepo: false, path: path}
	}
	ch <- IsGitRepoResult{IsGitRepo: isRepo, path: path}
}

// todo - shake duplicate results for same repo
func ScanDirsForGitRepos(path string, dirs []os.FileInfo) {
	out, err := exec.Command("git",  "-C", path, "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil && err.Error() != "exit status 128" {
		log.Fatal("Error while checking if working directory is git repo " + path + " ", err)
		return
	}
	isRepo, parseErr := common.ParseBoolFromBytes(out)
	if parseErr != nil {
		fmt.Println("Could not determine if working dir is git repo. Will continue with children...", parseErr)
	}
	if isRepo {
		fmt.Println("Working directory is git repo!", path)
		return
	}

	results := make(chan IsGitRepoResult)
	fmt.Print("\nStarting scan for git repositories...\n")
	var validRepos []string
	go runGitDirScans(path, dirs, results)
	for i := 0; i < len(dirs); i++ {
		processGitScanResult(results, &validRepos)
	}
	fmt.Println(validRepos)
}

func runGitDirScans(path string, dirs []os.FileInfo, results chan<- IsGitRepoResult) {
	for _, dir := range dirs {
		go IsDirGitRepo(path + "\\" + dir.Name(), results)
	}
}

func processGitScanResult(results <-chan IsGitRepoResult, outPutSlice *[]string) {
	scanResult := <-results
	if scanResult.IsGitRepo {
		*outPutSlice = append(*outPutSlice, scanResult.path)
		fmt.Println(scanResult.path + " is valid git repo")
	} else {
		fmt.Println(scanResult.path + " is not a git repo")
	}
}
