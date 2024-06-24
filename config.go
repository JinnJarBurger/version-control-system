package main

import (
	"bufio"
	"fmt"
	loggger "log"
	"os"
)

func processConfigArg(configFile *os.File) {
	config := func() string {
		if len(os.Args) > 2 {
			return os.Args[2]
		}

		return ""
	}()

	if config != "" {
		err := os.Truncate(configFile.Name(), 0)
		if err != nil {
			loggger.Fatal(err)
		}

		_, err = fmt.Fprintln(configFile, config)
		if err != nil {
			loggger.Fatal(err)
		}

		fmt.Printf("The username is %s.\n", config)

		return
	}

	scanner := bufio.NewScanner(configFile)
	scanner.Scan()

	if scanner.Text() != "" {
		fmt.Printf("The username is %s.\n", scanner.Text())
	} else {
		fmt.Println("Please, tell me who you are.")
	}
}
