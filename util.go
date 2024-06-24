package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	loggger "log"
	"os"
)

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
