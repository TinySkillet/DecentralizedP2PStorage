package main

import (
	"bytes"
	"testing"
)

func TestCopyEncyptDecrypt(t *testing.T) {

	payload := "Foo not bar"
	src := bytes.NewReader([]byte(payload))
	dest := new(bytes.Buffer)

	key := newEcryptionKey()
	_, err := copyEncrypt(key, src, dest)
	if err != nil {
		t.Error(err)
	}

	out := new(bytes.Buffer)
	nw, err := copyDecrypt(key, dest, out)
	if err != nil {
		t.Error(err)
	}

	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != payload {
		t.Errorf("Decryption failed!!")
	}
}
