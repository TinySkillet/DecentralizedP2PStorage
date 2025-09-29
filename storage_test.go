package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	key := "cooldawg"
	pathkey := CASPathTransformFunc(key)

	expectedOriginalKey := "1ff51b817f2aa0ff28845b648e54fa24e05cb151"
	expectedPathName := "1ff51/b817f/2aa0f/f2884/5b648/e54fa/24e05/cb151"

	assert.Equal(t, pathkey.Pathname, expectedPathName)
	assert.Equal(t, pathkey.Original, expectedOriginalKey)
}

// func TestStore(t *testing.T) {
// 	opts := StoreOpts{
// 		PathTransformFunc: CASPathTransformFunc,
// 	}

// 	s := NewStore(opts)

// 	data := bytes.NewReader([]byte("some jpeg"))

// 	assert.Nil(t, s.writeStream("myspecialpicture", data))

// }
