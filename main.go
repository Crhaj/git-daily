package main

import (
	"fmt"
	"git-daily/src/utils/cmd"
	"git-daily/src/utils/git"
)

var InitialWd = cmd.Pwd()

func main() {
	fmt.Println("************ Welcome to Git-Daily ************")
	fmt.Printf("Program was run from path: %v\n", InitialWd)
	//files := cmd.GetDirContent(InitialWd)
	// cmd.GetDirectories(files)
	git.StartCrawl(InitialWd)
}
