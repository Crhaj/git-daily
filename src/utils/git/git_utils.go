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

func StartCrawl(path string) {
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
		processSingleRepo(path)
	}
}

// todo - shake duplicate results for same repo
func ScanDirsForGitRepos(path string, dirs []os.FileInfo) {
	results := make(chan IsGitRepoResult)
	fmt.Print("\nStarting scan for git repositories...\n")
	var validRepos []string
	go runGitReposScans(path, dirs, results)
	for i := 0; i < len(dirs); i++ {
		processGitReposScanResult(results, &validRepos)
	}
	fmt.Println(validRepos)
}

func processSingleRepo(path string) {
	FetchPrune(path)
	initialBranchName := GetCurrentBranchName(path)
	shouldStash := HasUnstashedChanges(path)
	if shouldStash {
		Stash(path, false)
	}
	if initialBranchName != "master" {
		Checkout(path, "master", initialBranchName)
	}
	Pull(path)
	Checkout(path, initialBranchName, initialBranchName)
	if shouldStash {
		Stash(path, true)
	}
}

func FetchPrune(path string) {
	_, _ = runGitCommand("-C", path, "fetch", "--prune")
}

// todo - create and store named stash?
func Stash(path string, pop bool) {
	if pop {
		_, _ = runGitCommand("-C", path, "stash", "pop")
	} else {
		_, _ = runGitCommand("-C", path, "stash")
	}
}

// todo - return to init branch and pop if fails?
func Pull(path string) {
	_, err := runGitCommand("-C", path, "pull")
	if err != nil {
		fmt.Println("Could not pull origin", err)
	}
}

func Checkout(path string, branchName string, initialBranchName string) {
	_, err := runGitCommand("-C", path, "checkout", branchName)
	if err != nil {
		if GetCurrentBranchName(path) != initialBranchName && branchName != initialBranchName {
			Checkout(path, initialBranchName, initialBranchName)
		}
		// todo - more logs, pop stash if it was pushed to?
		log.Fatal("Error while checking out branch", branchName)
	}
}

func GetCurrentBranchName(path string) string {
	result, err := runGitCommand("-C", path, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		log.Fatal("Could not get current branch name, stopping...")
	}
	return result
}

func HasUnstashedChanges(path string) bool {
	result, err := runGitCommand("-C", path, "status", "--untracked-files=no", "--porcelain")
	if err != nil {
		log.Fatal("Error: cannot determine if repo has unstashed changes, stopping...")
	}
	return len(result) > 0
}

func runGitCommand(args... string) (string, error) {
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		fmt.Println("Error while running git command with args: ", args)
		fmt.Println(err)
		return "", err
	}
	parsedResult := common.ParseStringFromBytes(out)
	fmt.Println(parsedResult)
	return parsedResult, nil
}

func runGitReposScans(path string, dirs []os.FileInfo, results chan<- IsGitRepoResult) {
	for _, dir := range dirs {
		go IsDirGitRepo(path + "\\" + dir.Name(), results)
	}
}

func processGitReposScanResult(results <-chan IsGitRepoResult, outPutSlice *[]string) {
	scanResult := <-results
	if scanResult.IsGitRepo {
		*outPutSlice = append(*outPutSlice, scanResult.path)
		fmt.Println(scanResult.path + " is valid git repo")
	} else {
		fmt.Println(scanResult.path + " is not a git repo")
	}
}
