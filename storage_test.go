package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	key := "cooldawg"
	pathkey := CASPathTransformFunc(key)

	expectedFilename := "1ff51b817f2aa0ff28845b648e54fa24e05cb151"

	expectedPathname := "1ff51/b817f/2aa0f/f2884/5b648/e54fa/24e05/cb151"

	assert.Equal(t, pathkey.Filename, expectedFilename)
	assert.Equal(t, pathkey.Pathname, expectedPathname)
}

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "absolutechad"

	data := []byte("i don't fucking know bro")
	if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := range 50 {
		key := fmt.Sprintf("absolutechad_%d", i)
		data := []byte("some kind of png")

		if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		if !s.Has(key) {
			t.Errorf("Expected to have key %s\n", key)
		}

		assert.False(t, s.Has("tero bau"))

		_, r, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		b, _ := io.ReadAll(r)

		assert.Equal(t, b, data)
		assert.NoError(t, s.Delete(key))
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
