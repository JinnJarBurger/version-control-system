package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	loggger "log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		fmt.Println("Restore a file.")

	default:
		fmt.Printf("'%s' is not a SVCS command.", os.Args[1])

	}
}

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
				loggger.Fatal(err)
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

func processCommitArg(configFile, indexFile, logFile *os.File) {
	var filesToCommit []string
	var finalHash []byte
	var log strings.Builder

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
			loggger.Fatal(err)
		}

		// TODO: initially optimized space here, not needed for now but will bring this feature back later
		if dirEmpty || trackedFilesChanged {
			fileInIndexFile, err := openFile("./", filename)
			if err != nil {
				loggger.Fatal(err)
			}

			fileHash := md5.New()

			_, err = io.Copy(fileHash, fileInIndexFile)
			if err != nil {
				loggger.Fatal(err)
			}

			finalHash = xorMd5Hashes(md5Hash.Sum(nil), fileHash.Sum(nil))

			filesToCommit = append(filesToCommit, filename)

			closeFile(fileInIndexFile)
		}
	}

	if len(filesToCommit) > 0 {
		hashDir := "./vcs/commits/" + hex.EncodeToString(finalHash) + "/"

		err := os.MkdirAll(hashDir, os.ModePerm)
		if err != nil {
			loggger.Fatal(err)
		}

		for _, filename := range filesToCommit {
			file, err := openFile("./", filename)
			if err != nil {
				loggger.Fatal(err)
			}

			destFile, err := os.Create(hashDir + filename)
			if err != nil {
				loggger.Fatal(err)
			}

			_, err = io.Copy(destFile, file)
			if err != nil {
				loggger.Fatal(err)
			}

			closeFile(destFile)
			closeFile(file)
		}

		log.WriteString("commit ")
		log.WriteString(hex.EncodeToString(finalHash))
		log.WriteString("\n")

		scanner := bufio.NewScanner(configFile)
		scanner.Scan()

		log.WriteString("Author: ")
		log.WriteString(scanner.Text())
		log.WriteString("\n")

		log.WriteString(commit)

		tmpLogFile, err := os.Create("./vcs/tmpLog.txt")
		if err != nil {
			loggger.Fatal(err)
		}

		_, err = fmt.Fprintln(tmpLogFile, log.String())
		if err != nil {
			loggger.Fatal(err)
		}

		scanner = bufio.NewScanner(logFile)

		for scanner.Scan() {
			_, err = fmt.Fprintln(tmpLogFile, scanner.Text())
			if err != nil {
				loggger.Fatal(err)
			}

		}

		err = tmpLogFile.Sync()
		if err != nil {
			loggger.Fatal(err)
		}

		closeFile(tmpLogFile)

		err = os.Rename(tmpLogFile.Name(), logFile.Name())
		if err != nil {
			loggger.Fatal(err)
		}

		fmt.Println("Changes are committed.")
	} else {
		fmt.Println("Nothing to commit.")
	}

	//err := os.Truncate(indexFile.Name(), 0)
	//if err != nil {
	//	loggger.Fatal(err)
	//}
}

func anyFileChanged(indexFile *os.File, commitHash string, buffer *bytes.Buffer) bool {
	scanner := bufio.NewScanner(indexFile)

	for scanner.Scan() {
		filename := scanner.Text()

		_, err := fmt.Fprintln(buffer, filename)
		if err != nil {
			loggger.Fatal(err)
		}

		if fileChangedSinceLastCommit(filename, commitHash) {
			return true
		}
	}

	return false
}

func fileChangedSinceLastCommit(filename, commitHash string) bool {
	commitHashDir := "./vcs/commits/" + commitHash + "/"

	file, err := openFile("./", filename)
	if err != nil {
		loggger.Fatal(err)
	}
	defer closeFile(file)

	md5Hash1 := calculateMd5(file)

	entries, err := os.ReadDir(commitHashDir)
	if err != nil {
		loggger.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filename == entry.Name() {
			fileInfo, err := entry.Info()
			if err != nil {
				loggger.Fatal(err)
			}

			commitFile, err := openFile(commitHashDir, filepath.Base(fileInfo.Name()))
			if err != nil {
				loggger.Fatal(err)
			}

			md5Hash2 := calculateMd5(commitFile)

			closeFile(commitFile)

			return md5Hash1 != md5Hash2
		}
	}

	return false
}

func findLatestCommitDir() string {
	commitDir := "./vcs/commits/"

	entries, err := os.ReadDir(commitDir)
	if err != nil {
		loggger.Fatal(err)
	}

	var latestDirName string
	var latestCreatedTime int64

	for _, entry := range entries {
		if entry.IsDir() {
			fileInfo, err := entry.Info()
			if err != nil {
				loggger.Fatal(err)
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
	commitDir := "./vcs/commits/"

	dir, err := os.Open(commitDir)
	if err != nil {
		loggger.Fatal(err)
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

func calculateMd5(file *os.File) string {
	md5Hash := md5.New()
	_, err := io.Copy(md5Hash, file)
	if err != nil {
		loggger.Fatal(err)
	}

	return hex.EncodeToString(md5Hash.Sum(nil))
}

func xorMd5Hashes(hash1, hash2 []byte) []byte {
	if len(hash1) != len(hash2) {
		loggger.Fatal(errors.New("invalid hashes"))
	}

	result := make([]byte, len(hash1))

	for i := 0; i < len(hash1); i++ {
		result[i] = hash1[i] ^ hash2[i]
	}

	return result
}

func processHelpArg() {
	if len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] == "--help") {
		fmt.Printf(helpMsg)

		os.Exit(0)
	}
}

func createOrOpenFile(filepath, filename string) (*os.File, error) {
	if _, err := os.Stat(filepath + filename); errors.Is(err, os.ErrNotExist) {
		return os.Create(filepath + filename)
	}

	return os.OpenFile(filepath+filename, os.O_APPEND|os.O_RDWR, 0644)
}

func openFile(filepath, filename string) (*os.File, error) {
	return os.OpenFile(filepath+filename, os.O_APPEND|os.O_RDWR, 0644)
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		loggger.Fatal(err)
	}
}
