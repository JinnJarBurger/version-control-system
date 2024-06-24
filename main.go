package main

import (
	"fmt"
	loggger "log"
	"os"
)

const helpMsg = `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`

func main() {
	loggger.SetFlags(loggger.LstdFlags | loggger.Lshortfile)

	processHelpArg()

	err := os.MkdirAll("./vcs/commits/", os.ModePerm)
	if err != nil {
		loggger.Fatal(err)
	}

	configFile, err := createOrOpenFile("./vcs/", "config.txt")
	if err != nil {
		loggger.Fatal(err)
	}
	defer closeFile(configFile)

	indexFile, err := createOrOpenFile("./vcs/", "index.txt")
	if err != nil {
		loggger.Fatal(err)
	}
	defer closeFile(indexFile)

	logFile, err := createOrOpenFile("./vcs/", "log.txt")
	if err != nil {
		loggger.Fatal(err)
	}
	defer closeFile(logFile)

	switch os.Args[1] {
	case "config":
		processConfigArg(configFile)

	case "add":
		processAddArg(indexFile)

	case "log":
		processLogArg(logFile)

	case "commit":
		processCommitArg(configFile, indexFile, logFile)

	case "checkout":
		processCheckoutArg()

	default:
		fmt.Printf("'%s' is not a SVCS command.", os.Args[1])

	}
}

func processHelpArg() {
	if len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] == "--help") {
		fmt.Printf(helpMsg)

		os.Exit(0)
	}
}
