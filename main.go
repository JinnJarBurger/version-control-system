package main

import (
	"flag"
	"fmt"
	"os"
)

var supportedCommands = []string{"config", "add", "log", "commit", "checkout"}

func main() {
	var config, add, log, commit, checkout string

	flag.StringVar(&config, "config", "", "Get and set a username.")
	flag.StringVar(&add, "add", "", "Add a file to the index.")
	flag.StringVar(&log, "log", "", "Show commit logs.")
	flag.StringVar(&commit, "commit", "", "Save changes.")
	flag.StringVar(&checkout, "checkout", "", "Restore a file.")

	//flag.Parse()

	if len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] == "--help") {
		printDefault()

		os.Exit(0)
	}

	if flag.Lookup(os.Args[1]) == nil {
		fmt.Printf("'%s' is not a SVCS command.", os.Args[1])
	} else {
		fmt.Println(flag.Lookup(os.Args[1]).Usage)
	}
}

func printDefault() {
	fmt.Println("These are SVCS commands:")

	for _, command := range supportedCommands {
		fmt.Printf("%-10s %s\n", command, flag.Lookup(command).Usage)
	}
}
