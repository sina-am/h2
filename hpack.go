package main

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecodingNumber = errors.New("invalid byte for numeric representation")
)
var huffmanCodes = map[rune]string{
	'/':  "011000",
	'0':  "00000",
	'1':  "00001",
	'2':  "00010",
	'3':  "011001",
	'4':  "011010",
	'5':  "011011",
	'6':  "011100",
	'7':  "011101",
	'8':  "011110",
	'9':  "011111",
	':':  "1011100",
	';':  "11111011",
	'<':  "111111111111100",
	'=':  "100000",
	'>':  "111111111011",
	'?':  "1111111100",
	'@':  "1111111111010",
	'A':  "100001",
	'B':  "1011101",
	'C':  "1011110",
	'D':  "1011111",
	'E':  "1100000",
	'F':  "1100001",
	'G':  "1100010",
	'H':  "1100011",
	'I':  "1100100",
	'J':  "1100101",
	'K':  "1100110",
	'L':  "1100111",
	'M':  "1101000",
	'N':  "1101001",
	'O':  "1101010",
	'P':  "1101011",
	'Q':  "1101100",
	'R':  "1101101",
	'S':  "1101110",
	'T':  "1101111",
	'U':  "1110000",
	'V':  "1110001",
	'W':  "1110010",
	'X':  "11111100",
	'Y':  "1110011",
	'Z':  "11111101",
	'[':  "1111111111011",
	'\\': "1111111111111110|000",
	']':  "1111111111100",
	'^':  "11111111111100",
	'_':  "100010",
	'`':  "11111111|1111101",
	'a':  "00011",
	'b':  "100011",
	'c':  "00100",
	'd':  "100100",
	'e':  "00101",
	'f':  "100101",
	'g':  "100110",
	'h':  "100111",
	'i':  "00110",
	'j':  "1110100",
	'k':  "1110101",
	'l':  "101000",
	'm':  "101001",
	'n':  "101010",
	'o':  "00111",
	'p':  "101011",
	'q':  "1110110",
	'r':  "101100",
	's':  "01000",
	't':  "01001",
	'u':  "101101",
	'v':  "1110111",
	'w':  "1111000",
	'x':  "1111001",
	'y':  "1111010",
	'z':  "1111011",
	'{':  "111111111111110",
	'|':  "11111111100",
	'}':  "11111111111101",
	'~':  "1111111111101",
}

func HuffmanEncodeString(s string) []byte {
	seq := ""
	for _, char := range s {
		seq += huffmanCodes[char]
	}

	paddingLength := 8 - (len(seq) % 8)
	padding := ""
	for i := 0; i < paddingLength; i++ {
		padding += "0"
	}
	seq = padding + seq

	b := []byte{}
	for i := 0; i < len(seq); i += 8 {
		var n uint64 = 0
		for j := 0; j < 8; j++ {
			if seq[i+j] == '1' {
				n += pow(2, uint8(7-j))
			}
		}

		b = append(b, uint8(n))
	}

	return b
}

type HeaderField struct {
	name  string
	value string
}

var staticTable = []HeaderField{
	{name: "", value: ""},
	{name: ":authority:", value: ""},
	{name: ":method:", value: "GET"},
	{name: ":method:", value: "POST"},
	{name: ":path:", value: "/"},
	{name: ":path:", value: "/index.html"},
	{name: ":scheme:", value: "http"},
	{name: ":scheme:", value: "https"},
	{name: ":status:", value: "200"},
	{name: ":status:", value: "404"},
	{name: ":status:", value: "500"},
	{name: "accept-charset", value: ""},
	{name: "accept-encoding", value: ""},
	{name: "accept-language", value: ""},
	{name: "accept-ranges", value: ""},
	{name: "accept", value: ""},
	{name: "access-control-allow-origin", value: ""},
	{name: "age", value: ""},
	{name: "allow", value: ""},
	{name: "authorization", value: ""},
	{name: "cache-control", value: ""},
	{name: "content-disposition", value: ""},
	{name: "content-encoding", value: ""},
	{name: "content-language", value: ""},
	{name: "content-length", value: ""},
	{name: "content-location", value: ""},
	{name: "content-range", value: ""},
	{name: "content-type", value: ""},
	{name: "cookie", value: ""},
	{name: "date", value: ""},
	{name: "etag", value: ""},
	{name: "expect", value: ""},
	{name: "expires", value: ""},
	{name: "from", value: ""},
	{name: "host", value: ""},
	{name: "if-match", value: ""},
	{name: "if-modified-since", value: ""},
	{name: "if-none-match", value: ""},
	{name: "if-range", value: ""},
	{name: "if-unmodified-since", value: ""},
	{name: "last-modified", value: ""},
	{name: "link", value: ""},
	{name: "location", value: ""},
	{name: "max-forwards", value: ""},
	{name: "proxy-authenticate", value: ""},
	{name: "proxy-authorization", value: ""},
	{name: "range", value: ""},
	{name: "referer", value: ""},
	{name: "refresh", value: ""},
	{name: "retry-after", value: ""},
	{name: "server", value: ""},
	{name: "set-cookie", value: ""},
	{name: "strict-transport-security", value: ""},
	{name: "transfer-encoding", value: ""},
	{name: "user-agent", value: ""},
	{name: "vary", value: ""},
	{name: "via", value: ""},
	{name: "www-authenticate", value: ""},
}

func pow(base uint64, to uint8) uint64 {
	var (
		i      uint8  = 0
		result uint64 = 1
	)
	for ; i < to; i++ {
		result *= base
	}

	return result
}

func decodeInteger(b []byte, prefix uint8) (uint64, error) {
	prefixValue := pow(2, prefix) - 1

	if uint64(b[0]) < prefixValue {
		return uint64(b[0]), nil
	}
	if len(b) < 2 {
		return 0, ErrDecodingNumber
	}

	var (
		n        uint64
		m        uint8
		nextByte uint64
		i        = uint8(1)
	)
	for {
		nextByte = uint64(b[i])
		n += (nextByte & 127) * (1 << m)

		if ((nextByte & 128) >> 7) == 0 {
			break
		}
		m += 7
		i++
	}

	return (n + prefixValue), nil
}

func encodeInteger(n uint64, prefix uint8) []byte {
	prefixValue := pow(2, prefix) - 1

	if n < prefixValue {
		return []byte{uint8(n)}
	}

	bytes := []byte{uint8(prefixValue)}
	n = n - prefixValue
	for n >= 128 {
		bytes = append(bytes, uint8((n%128)+128))
		n = n / 128
	}
	bytes = append(bytes, uint8(n))
	return bytes
}

func encodeStringLiteral(s string, huffmanEncoded bool) []byte {
	length := encodeInteger(uint64(len(s)), 7)
	if huffmanEncoded {
		length[0] |= 128
		return append(length, HuffmanEncodeString(s)...)
	}

	length[0] &= 0x7F
	return append(length, []byte(s)...)
}

func decodeStringLiteral(b []byte, huffmanEncoded bool) (string, error) {
	if huffmanEncoded {
		panic("huffman decoding is not implemented yet")
	}

	length, err := decodeInteger(b, 7)
	if err != nil {
		return "", err
	}

	return string(b[1 : length+1]), nil
}

type HPackEncoder interface {
	Encode(writer io.Writer, headerFields []HeaderField) (int, error)
}

type hPackEncoder struct {
	dynamicTable     []HeaderField
	isHuffmanEncoded bool
}

func NewHPackEncoder() HPackEncoder {
	return &hPackEncoder{
		dynamicTable:     []HeaderField{},
		isHuffmanEncoded: true,
	}
}

func (h *hPackEncoder) encodeHeaderField(headerField HeaderField, indexed bool, newName bool) []byte {
	indexedHeaderField := 0
	indexedHeaderName := 0
	for index, item := range staticTable {
		if headerField.name == item.name {
			if headerField.value == item.value {
				indexedHeaderField = index
				break
			}
			if indexedHeaderName != 0 {
				break
			}
			indexedHeaderName = index
		}
	}
	// Indexed Header Field Representation
	if indexedHeaderField != 0 {
		bytes := encodeInteger(uint64(indexedHeaderField), 7)
		bytes[0] |= 0x80
		return bytes
	}

	// Literal Header Field with Incremental Indexing -- Indexed name
	if indexedHeaderName != 0 && indexed && !newName {
		bytes := encodeInteger(uint64(indexedHeaderName), 6)
		bytes[0] |= 0x40

		return append(bytes, encodeStringLiteral(headerField.value, h.isHuffmanEncoded)...)
	}
	// Literal Header Field with Incremental Indexing -- New Name
	if indexedHeaderName != 0 && indexed && newName {
		bytes := []byte{0x40}

		bytes = append(bytes, encodeStringLiteral(headerField.name, h.isHuffmanEncoded)...)
		return append(bytes, encodeStringLiteral(headerField.value, h.isHuffmanEncoded)...)
	}

	// Literal Header Field without Indexing -- Indexed name
	if indexedHeaderName != 0 && !indexed {
		bytes := encodeInteger(uint64(indexedHeaderName), 4)
		bytes[0] &= 0x0f
		return append(bytes, encodeStringLiteral(headerField.value, h.isHuffmanEncoded)...)
	}

	fmt.Printf("%s: %s\n", headerField.name, headerField.value)
	panic("not implemented yet")
}

func (h *hPackEncoder) Encode(writer io.Writer, headerFields []HeaderField) (int, error) {
	// takes an ordered header list and encode it
	total_bytes := 0
	for _, headerField := range headerFields {
		n, err := writer.Write(h.encodeHeaderField(headerField, false, false))
		if err != nil {
			return total_bytes, err
		}
		total_bytes += n
	}

	return total_bytes, nil
}
