package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var commitDir = filepath.Join(".", "vcs", "commits")

func processCommitArg(configFile, indexFile, logFile *os.File) {
	var filesToCommit []string
	var finalHash []byte
	var logBuilder strings.Builder

	md5Hash := md5.New()
	md5Hash.Write([]byte(time.Now().String()))

	commit := func() string {
		if len(os.Args) > 2 {
			return strings.ReplaceAll(os.Args[2], "\"", "")
		}

		return ""
	}()

	if commit == "" {
		fmt.Println("Message was not passed.")

		return
	}

	if fileInfo, _ := indexFile.Stat(); fileInfo.Size() == 0 {
		fmt.Println("Nothing to commit.")

		return
	}

	latestCommitDir := findLatestCommitDir()
	bufferForFileName := bytes.Buffer{}
	trackedFilesChanged := anyFileChanged(indexFile, latestCommitDir, &bufferForFileName)
	scanner := bufio.NewScanner(&bufferForFileName)

	for scanner.Scan() {
		filename := scanner.Text()

		dirEmpty, err := commitDirEmpty()
		if err != nil {
			log.Fatal(err)
		}

		// TODO: initially optimized space here, not needed for now but will bring this feature back later
		if dirEmpty || trackedFilesChanged {
			fileInIndexFile, err := openFile(".", filename)
			if err != nil {
				log.Fatal(err)
			}

			fileHash := md5.New()

			_, err = io.Copy(fileHash, fileInIndexFile)
			if err != nil {
				log.Fatal(err)
			}

			finalHash = xorMd5Hashes(md5Hash.Sum(nil), fileHash.Sum(nil))

			filesToCommit = append(filesToCommit, filename)

			closeFile(fileInIndexFile)
		}
	}

	if len(filesToCommit) > 0 {
		hashDir := filepath.Join(commitDir, hex.EncodeToString(finalHash))

		err := os.MkdirAll(hashDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		for _, filename := range filesToCommit {
			file, err := openFile(".", filename)
			if err != nil {
				log.Fatal(err)
			}

			destFile, err := os.Create(filepath.Join(hashDir, filename))
			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(destFile, file)
			if err != nil {
				log.Fatal(err)
			}

			closeFile(destFile)
			closeFile(file)
		}

		logBuilder.WriteString("commit ")
		logBuilder.WriteString(hex.EncodeToString(finalHash))
		logBuilder.WriteString("\n")

		scanner := bufio.NewScanner(configFile)
		scanner.Scan()

		logBuilder.WriteString("Author: ")
		logBuilder.WriteString(scanner.Text())
		logBuilder.WriteString("\n")

		logBuilder.WriteString(commit)
		logBuilder.WriteString("\n")

		tmpLogFile, err := os.Create(filepath.Join(".", "vcs", "tmpLog.txt"))
		if err != nil {
			log.Fatal(err)
		}

		_, err = fmt.Fprintln(tmpLogFile, logBuilder.String())
		if err != nil {
			log.Fatal(err)
		}

		scanner = bufio.NewScanner(logFile)

		for scanner.Scan() {
			_, err = fmt.Fprintln(tmpLogFile, scanner.Text())
			if err != nil {
				log.Fatal(err)
			}

		}

		err = tmpLogFile.Sync()
		if err != nil {
			log.Fatal(err)
		}

		closeFile(tmpLogFile)

		err = os.Rename(tmpLogFile.Name(), logFile.Name())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Changes are committed.")
	} else {
		fmt.Println("Nothing to commit.")
	}
}

func anyFileChanged(indexFile *os.File, commitHash string, buffer *bytes.Buffer) bool {
	scanner := bufio.NewScanner(indexFile)

	for scanner.Scan() {
		filename := scanner.Text()

		_, err := fmt.Fprintln(buffer, filename)
		if err != nil {
			log.Fatal(err)
		}

		if fileChangedSinceLastCommit(filename, commitHash) {
			return true
		}
	}

	return false
}

func fileChangedSinceLastCommit(filename, commitHash string) bool {
	commitHashDir := filepath.Join(commitDir, commitHash)

	file, err := openFile(".", filename)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)

	md5Hash1 := calculateMd5(file)

	entries, err := os.ReadDir(commitHashDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filename == entry.Name() {
			fileInfo, err := entry.Info()
			if err != nil {
				log.Fatal(err)
			}

			commitFile, err := openFile(commitHashDir, filepath.Base(fileInfo.Name()))
			if err != nil {
				log.Fatal(err)
			}

			md5Hash2 := calculateMd5(commitFile)

			closeFile(commitFile)

			return md5Hash1 != md5Hash2
		}
	}

	return false
}

func findLatestCommitDir() string {
	entries, err := os.ReadDir(commitDir)
	if err != nil {
		log.Fatal(err)
	}

	var latestDirName string
	var latestCreatedTime int64

	for _, entry := range entries {
		if entry.IsDir() {
			fileInfo, err := entry.Info()
			if err != nil {
				log.Fatal(err)
			}

			createdTime := fileInfo.ModTime().Unix()

			if createdTime > latestCreatedTime {
				latestCreatedTime = createdTime
				latestDirName = entry.Name()
			}
		}
	}

	return latestDirName
}

func commitDirEmpty() (bool, error) {
	dir, err := os.Open(commitDir)
	if err != nil {
		log.Fatal(err)
	}

	defer closeFile(dir)

	_, err = dir.Readdirnames(1)

	if errors.Is(err, io.EOF) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return false, nil
}
