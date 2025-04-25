package main

import (
	"bytes"
	"testing"
)

func TestNumericRepresentation(t *testing.T) {
	prefix := uint8(5)
	for n := uint64(0); n < 100_000; n++ {
		n2, err := decodeInteger(encodeInteger(n, prefix), prefix)
		if err != nil {
			t.Errorf("decoding error: %s", err)
		}
		if n2 != n {
			t.Errorf("Excepted %d got %d", n, n2)
		}
	}
}

func TestHeaderFieldEncoding(t *testing.T) {
	expected := []byte{
		0x82, 0x04, 0x84, 0x61,
		0x25, 0x42, 0x7f, 0x87,
		0x41, 0x86, 0xa0, 0xe4,
		0x1d, 0x13, 0x9d, 0x09,
		0x7a, 0x88, 0x25, 0xb6,
		0x50, 0xc3, 0xab, 0xbc,
		0xda, 0xe0, 0x53, 0x03,
		0x2a, 0x2f, 0x2a,
	}

	headerFields := []HeaderField{
		{name: ":method:", value: "GET"},
		{name: ":path:", value: "/test"},
		{name: ":scheme:", value: "https"},
		{name: ":authority:", value: "localhost"},
		{name: "user-agent", value: "curl/7.85.0"},
		{name: "accept", value: "*/*"},
	}

	encoder := NewHPackEncoder()
	writer := bytes.Buffer{}
	if _, err := encoder.Encode(&writer, headerFields); err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expected, writer.Bytes()) {
		t.Errorf("expected %v got %v", expected, writer.Bytes())
	}
}
