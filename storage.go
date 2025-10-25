package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"
)

const DEFAULT_ROOT_FOLDER = "p2pnetwork"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashString) / blockSize

	paths := make([]string, sliceLen)

	for i := range sliceLen {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashString[from:to]
	}

	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashString,
	}
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) Delete(key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleted [%s] from disk", pathKey.Filename)
	}()

	parentDir := strings.Split(pathKey.FullPath(), "/")[0]

	parentDirWRoot := fmt.Sprintf("%s/%s", s.Root, parentDir)
	if err := os.RemoveAll(parentDirWRoot); err != nil {
		return err
	}
	return nil
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Has(key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPath := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	_, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false
		}
	}
	return true
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)

	fullPath := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(fullPath)
}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {

	pathKey := s.PathTransformFunc(key)

	pathnameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.Pathname)
	if err := os.MkdirAll(pathnameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}

	fullPath := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	f, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}
	return n, nil
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

func DefaultPathTransformFunc(key string) PathKey {
	return PathKey{
		Pathname: key,
		Filename: key,
	}
}

type Store struct {
	StoreOpts
}

type StoreOpts struct {
	// Root is the root folder containing all files and folders of the p2p system
	Root              string
	PathTransformFunc PathTransformFunc
}

func NewStore(opts StoreOpts) *Store {

	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = DEFAULT_ROOT_FOLDER
	}
	return &Store{
		StoreOpts: opts,
	}
}
