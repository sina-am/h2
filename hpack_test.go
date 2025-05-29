package main

import (
	"bytes"
	"fmt"
	"testing"
)

func bytesRepresentation(b []byte) string {
	out := ""
	for i := 0; i < len(b); i++ {
		out += fmt.Sprintf("%x ", b[i])
	}
	return out
}

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
		t.Errorf("\r\nexp %s\r\ngot %s", bytesRepresentation(expected), bytesRepresentation(writer.Bytes()))
	}
}
func TestHeaderFieldDecoding(t *testing.T) {
	expectedHeaders := []HeaderField{
		{name: ":status:", value: "200"},
		{name: "server", value: "nginx/1.24.0"},
		{name: "date", value: "Fri, 23 May 2025 16:12:32 GMT"},
		{name: "content-type", value: "text/html"},
		{name: "content-length", value: "8474"},
		{name: "last-modified", value: "Mon, 28 Mar 2022 19:46:48 GMT"},
		{name: "etag", value: "\"624210a8-211a\""},
		{name: "accept-ranges", value: "bytes"},
	}

	rawHeaders := bytes.NewReader([]byte{
		0x88, 0x76, 0x89, 0xaa, 0x63, 0x55, 0xe5, 0x80, 0xae, 0x26, 0x97, 0x07,
		0x61, 0x96, 0xc3, 0x61, 0xbe, 0x94, 0x13, 0x2a, 0x68, 0x1f, 0xa5, 0x04,
		0x01, 0x36, 0xa0, 0x5c, 0xb8, 0x11, 0x5c, 0x64, 0x4a, 0x62, 0xd1, 0xbf,
		0x5f, 0x87, 0x49, 0x7c, 0xa5, 0x89, 0xd3, 0x4d, 0x1f, 0x5c, 0x04, 0x38,
		0x34, 0x37, 0x34, 0x6c, 0x96, 0xd0, 0x7a, 0xbe, 0x94, 0x13, 0xca, 0x68,
		0x1d, 0x8a, 0x08, 0x02, 0x12, 0x81, 0x7e, 0xe3, 0x4e, 0x5c, 0x69, 0xe5,
		0x31, 0x68, 0xdf, 0x00, 0x83, 0x2a, 0x47, 0x37, 0x8c, 0xfe, 0x5c, 0x13,
		0x42, 0x08, 0x06, 0xf2, 0xc2, 0x08, 0x47, 0xfc, 0xff, 0x00, 0x89, 0x19,
		0x08, 0x5a, 0xd2, 0xb5, 0x83, 0xaa, 0x62, 0xa3, 0x84, 0x8f, 0xd2, 0x4a,
		0x8f,
	})

	headers := []HeaderField{}
	decoder := NewHPackDecoder()
	if err := decoder.Decode(rawHeaders, &headers); err != nil {
		t.Errorf("got error %s", err)
	}

	if len(headers) != len(expectedHeaders) {
		t.Errorf("not enough headers: expected %d got %d", len(expectedHeaders), len(headers))
		return
	}

	for i := 0; i < len(headers); i++ {
		if headers[i].name != expectedHeaders[i].name {
			t.Errorf("invalid header name: expected %s got %s", expectedHeaders[i].name, headers[i].name)
		}
		if headers[i].value != expectedHeaders[i].value {
			t.Errorf("invalid header value: expected %s got %s", expectedHeaders[i].value, headers[i].value)
		}
	}
}
