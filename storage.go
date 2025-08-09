package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

func CASPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashString) / blockSize

	paths := make([]string, sliceLen)

	for i := range sliceLen {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashString[from:to]
	}

	return strings.Join(paths, "/")
}

type PathTransformFunc func(string) string

func DefaultPathTransformFunc(key string) string {
	return key
}

type Store struct {
	StoreOpts
}

type StoreOpts struct {
	PathTransformFunc PathTransformFunc
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathName := s.PathTransformFunc(key)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	filename := "somefilename"
	pathAndFileName := pathName + "/" + filename

	f, err := os.Create(pathAndFileName)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written %d bytes to disk: %s\n", n, pathAndFileName)
	return nil
}
