package main

import (
	"bufio"
	"fmt"
	"os"
)

func processLogArg(logFile *os.File) {
	if fileInfo, _ := logFile.Stat(); fileInfo.Size() == 0 {
		fmt.Println("No commits yet.")

		return
	}

	scanner := bufio.NewScanner(logFile)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
