package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func processCheckoutArg() {
	checkout := func() string {
		if len(os.Args) > 2 {
			return os.Args[2]
		}

		return ""
	}()

	if checkout == "" {
		fmt.Println("Commit id was not passed.")

		return
	}

	if !commitExists(checkout) {
		fmt.Println("Commit does not exist.")

		return
	}

	commitHashDir := filepath.Join(commitDir, checkout)
	entries, err := os.ReadDir(commitHashDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			trackedFile, err := openFile(".", entry.Name())
			if err != nil {
				log.Fatal(err)
			}

			commitFile, err := openFile(commitHashDir, entry.Name())
			if err != nil {
				log.Fatal(err)
			}

			err = os.Truncate(trackedFile.Name(), 0)
			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(trackedFile, commitFile)
			if err != nil {
				log.Fatal(err)
			}

			closeFile(trackedFile)
			closeFile(commitFile)
		}
	}

	fmt.Printf("Switched to commit %s.\n", checkout)
}

func commitExists(commitHash string) bool {
	entries, err := os.ReadDir(commitDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == commitHash {
			return true
		}
	}

	return false
}
