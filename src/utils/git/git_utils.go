package git

import (
	"fmt"
	"git-daily/src/utils/cmd"
	"git-daily/src/utils/common"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

const MasterBranch = "master"

var wg sync.WaitGroup

type IsGitRepoResult struct {
	IsGitRepo bool
	path      string
}

func IsDirGitRepo(path string, ch chan<- IsGitRepoResult) {
	result, err := runGitCommand(path, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		ch <- IsGitRepoResult{IsGitRepo: false, path: path}
		return
	}
	isRepo, parseErr := strconv.ParseBool(result)
	if parseErr != nil {
		fmt.Println("Error parsing check result for:", path, parseErr)
		ch <- IsGitRepoResult{IsGitRepo: false, path: path}
	}
	ch <- IsGitRepoResult{IsGitRepo: isRepo, path: path}
}

func StartCrawl(path string) {
	result, err := runGitCommand(path, "rev-parse", "--is-inside-work-tree")
	if err != nil && err.Error() != "exit status 128" {
		log.Fatal("Error while checking if working directory is git repo " + path + " ", err)
		return
	}
	isRepo, parseErr := strconv.ParseBool(result)
	if parseErr != nil {
		fmt.Println("Could not determine if working dir is git repo. Continuing with children...", parseErr)
	}
	if isRepo {
		fmt.Println("Working directory is git repo!", path)
		processSingleRepo(path, false)
	} else {
		fmt.Println("Working directory is not a git repo, crawling children...")
		files := cmd.GetDirContent(path)
		pathsToRepos := ScanDirsForGitRepos(path, cmd.GetDirectories(files))
		if len(pathsToRepos) == 0 {
			fmt.Println("No git repositories found! Ending execution.")
			return
		}
		wg.Add(len(pathsToRepos))
		for _, path := range pathsToRepos {
			go processSingleRepo(path, true)
		}
		wg.Wait()
		return
	}
}

func ScanDirsForGitRepos(path string, dirs []os.FileInfo) []string {
	results := make(chan IsGitRepoResult)
	fmt.Print("\nStarting scan for git repositories...\n")
	var validRepos []string
	go runGitReposScans(path, dirs, results)
	for i := 0; i < len(dirs); i++ {
		processGitReposScanResult(results, &validRepos)
	}
	return validRepos
}

func processSingleRepo(path string, shouldRemoveFromWg bool) {
	defer func() {
		if shouldRemoveFromWg {
			wg.Done()
		}
	}()

	FetchPrune(path)
	initialBranchName := GetCurrentBranchName(path)
	shouldStash := HasUnstashedChanges(path)
	if shouldStash {
		Stash(path, false)
	}
	if initialBranchName != MasterBranch {
		if initialBranchName == "" {
			fmt.Println("Failed to get current branch name, skipping update of repo")
			if shouldStash {
				Stash(path, true)
			}
			return
		}
		Checkout(path, MasterBranch, initialBranchName)
	}
	Pull(path)
	Checkout(path, initialBranchName, initialBranchName)
	if shouldStash {
		Stash(path, true)
	}
}

func FetchPrune(path string) {
	_, _ = runGitCommand(path, "fetch", "--prune")
}

// todo - create and store named stash?
func Stash(path string, pop bool) {
	if pop {
		_, _ = runGitCommand(path, "stash", "pop")
	} else {
		_, _ = runGitCommand(path, "stash")
	}
}

// todo - return to init branch and pop if fails?
func Pull(path string) {
	_, err := runGitCommand(path, "pull")
	if err != nil {
		fmt.Println("Could not pull origin", err)
	}
}

func Checkout(path string, branchName string, initialBranchName string) {
	_, err := runGitCommand(path, "checkout", branchName)
	if err != nil {
		if GetCurrentBranchName(path) != initialBranchName && branchName != initialBranchName {
			Checkout(path, initialBranchName, initialBranchName)
		}
		// todo - more logs, pop stash if it was pushed to?
		log.Fatal("Error while checking out branch", branchName)
	}
}

func GetCurrentBranchName(path string) string {
	result, err := runGitCommand(path, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		fmt.Println("Could not get current branch name", err)
		return ""
	}
	return result
}

func HasUnstashedChanges(path string) bool {
	result, err := runGitCommand(path, "status", "--untracked-files=no", "--porcelain")
	if err != nil {
		log.Fatal("Error: cannot determine if repo has unstashed changes, stopping...")
	}
	return len(result) > 0
}

func runGitCommand(path string, args... string) (string, error) {
	params := append([]string{"-C", path}, args...)
	out, err := exec.Command("git", params...).CombinedOutput()
	if err != nil {
		fmt.Println("Error in:", append([]string{"git"}, params...), err)
		return "", err
	}
	parsedResult := common.ParseStringFromBytes(out)
	fmt.Println("Running:", append([]string{"git"}, params...))
	if parsedResult != "" {
		fmt.Println(parsedResult)
	} else {
		fmt.Println("Ok!")
	}
	return parsedResult, nil
}

func runGitReposScans(path string, dirs []os.FileInfo, results chan<- IsGitRepoResult) {
	for _, dir := range dirs {
		go IsDirGitRepo(filepath.FromSlash(path + string(os.PathSeparator) + dir.Name()), results)
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
