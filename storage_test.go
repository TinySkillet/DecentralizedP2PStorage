package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformFunc(t *testing.T) {
	key := "cooldawg"
	pathname := CASPathTransformFunc(key)
	fmt.Println(pathname)
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}

	s := NewStore(opts)

	data := bytes.NewReader([]byte("some jpeg"))

	assert.Nil(t, s.writeStream("myspecialpicture", data))

}
