package main

import (
	"fmt"
	"git-daily/src/utils/cmd"
	"git-daily/src/utils/git"
)

var InitialWd = cmd.Pwd()

const initMsg = `***********************************************
             Welcome to git-daily              
***********************************************`

const successMsg = `
***********************************************
    Your Master branches are now up to date    
***********************************************`

func main() {
	fmt.Println(initMsg)
	fmt.Printf("Program was run from path: %v\n", InitialWd)
	git.StartCrawl(InitialWd)
	fmt.Println(successMsg)
}
