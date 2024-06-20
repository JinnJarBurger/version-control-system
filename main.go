package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	loggger "log"
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

	processArgs()
	processInvalidArg()

	err := os.MkdirAll("./vcs", os.ModePerm)
	logError(err)

	configFile, err := createOrOpenFile("config.txt")
	logError(err)
	defer closeFile(configFile)

	indexFile, err := createOrOpenFile("index.txt")
	logError(err)
	defer closeFile(indexFile)

	processConfigFlag(config, configFile)
	processAddFlag(add, indexFile)
}

func processInvalidArg() {
	if os.Args[1] == "config" || os.Args[1] == "add" {
		return
	}

	if flag.Lookup(os.Args[1]) == nil {
		fmt.Printf("'%s' is not a SVCS command.", os.Args[1])
	} else {
		fmt.Println(flag.Lookup(os.Args[1]).Usage)
	}
}

func processConfigFlag(config string, configFile *os.File) {
	if os.Args[1] != "config" {
		return
	}

	if len(os.Args) > 2 {
		config = os.Args[2]
	}

	if config != "" {
		err := os.Truncate(configFile.Name(), 0)
		logError(err)

		_, err = fmt.Fprintln(configFile, config)
		logError(err)

		fmt.Printf("The username is %s.\n", config)
	} else {
		scanner := bufio.NewScanner(configFile)
		scanner.Scan()

		if scanner.Text() != "" {
			fmt.Printf("The username is %s.\n", scanner.Text())
		} else {
			fmt.Println("Please, tell me who you are.")
		}
	}
}

func processAddFlag(add string, indexFile *os.File) {
	if os.Args[1] != "add" {
		return
	}

	if len(os.Args) > 2 {
		add = os.Args[2]
	}

	if add != "" {
		if _, err := os.Stat(add); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Can't find '%s'.", add)
		} else {
			_, err := fmt.Fprintln(indexFile, add)
			logError(err)

			fmt.Printf("The file '%s' is tracked.", add)
		}
	} else if fileInfo, _ := indexFile.Stat(); fileInfo.Size() > 0 {
		scanner := bufio.NewScanner(indexFile)

		fmt.Println("Tracked files:")

		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	} else {
		fmt.Println("Add a file to the index.")
	}
}

func processArgs() {
	if len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] == "--help") {
		printDefault()

		os.Exit(0)
	} else {
		flag.Parse()
	}
}

func printDefault() {
	fmt.Println("These are SVCS commands:")

	for _, command := range supportedCommands {
		fmt.Printf("%-10s %s\n", command, flag.Lookup(command).Usage)
	}
}

func isFlagPassed(name string) bool {
	found := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}

func createOrOpenFile(filename string) (*os.File, error) {
	if _, err := os.Stat("./vcs/" + filename); errors.Is(err, os.ErrNotExist) {
		return os.Create("./vcs/" + filename)
	}

	return os.OpenFile("./vcs/"+filename, os.O_APPEND|os.O_RDWR, 0644)
}

func closeFile(file *os.File) {
	err := file.Close()
	logError(err)
}

func logError(err error) {
	if err != nil {
		loggger.Fatal(err)
	}
}
