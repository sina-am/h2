package main

import (
	"bytes"
	"testing"
)

func TestHuffmanEncode(t *testing.T) {
	expectedEncoded := []byte{170, 99, 85, 229, 128, 174, 38, 151, 7}
	s := "nginx/1.24.0"
	encoded := HuffmanEncode(s)

	if !bytes.Equal(expectedEncoded, encoded) {
		t.Errorf("expected %s got %s", expectedEncoded, encoded)
	}
}
func TestHuffmanDecode(t *testing.T) {
	rawEncoded := []byte{170, 99, 85, 229, 128, 174, 38, 151, 7}
	expectedDecoded := "nginx/1.24.0"

	decoded := HuffmanDecode(rawEncoded)

	if expectedDecoded != decoded {
		t.Errorf("expected %s got %s", expectedDecoded, decoded)
	}
}
