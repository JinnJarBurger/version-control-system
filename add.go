package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
)

func processAddArg(indexFile *os.File) {
	add := func() string {
		if len(os.Args) > 2 {
			return os.Args[2]
		}

		return ""
	}()

	if add != "" {
		if _, err := os.Stat(add); errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Can't find '%s'.\n", add)
		} else {
			_, err := fmt.Fprintln(indexFile, add)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("The file '%s' is tracked.\n", add)
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
