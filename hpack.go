package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
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
	'.':  "010111",
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

var huffmanCodes2 = map[rune]string{
	0x00: "1111111111000",
	0x01: "11111111111111111011000",
	0x02: "1111111111111111111111100010",
	0x03: "1111111111111111111111100011",
	0x04: "1111111111111111111111100100",
	0x05: "1111111111111111111111100101",
	0x06: "1111111111111111111111100110",
	0x07: "1111111111111111111111100111",
	0x08: "1111111111111111111111101000",
	0x09: "111111111111111111101010",
	0x0A: "111111111111111111111111111100",
	0x0B: "1111111111111111111111101001",
	0x0C: "1111111111111111111111101010",
	0x0D: "1111111111111111111111101011",
	0x0E: "1111111111111111111111101100",
	0x0F: "1111111111111111111111101101",
	0x10: "1111111111111111111111101110",
	0x11: "1111111111111111111111101111",
	0x12: "1111111111111111111111110000",
	0x13: "1111111111111111111111110001",
	0x14: "1111111111111111111111110010",
	0x15: "111111111111111111111111111101",
	0x16: "1111111111111111111111110011",
	0x17: "1111111111111111111111110100",
	0x18: "1111111111111111111111110101",
	0x19: "1111111111111111111111110110",
	0x1A: "1111111111111111111111110111",
	0x1B: "1111111111111111111111111000",
	0x1C: "1111111111111111111111111001",
	0x1D: "1111111111111111111111111010",
	0x1E: "1111111111111111111111111011",
	0x1F: "1111111111111111111111111100",
	' ':  "010100",
	'!':  "1111111000",
	'"':  "1111111001",
	'#':  "111111111010",
	'$':  "1111111111001",
	'%':  "010101",
	'&':  "11111000",
	'\'': "11111111010",
	'(':  "1111111010",
	')':  "1111111011",
	'*':  "11111001",
	'+':  "11111111011",
	',':  "11111010",
	'-':  "010110",
	'.':  "010111",
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
	'X':  "111111111111101",
	'Y':  "1110011",
	'Z':  "1110100",
	'[':  "1111111111011",
	'\\': "111111111111110",
	']':  "1111111111100",
	'^':  "1111111111111110000",
	'_':  "100010",
	'`':  "1111111111111110001",
	'a':  "00011",
	'b':  "100011",
	'c':  "00100",
	'd':  "100100",
	'e':  "00101",
	'f':  "100101",
	'g':  "100110",
	'h':  "100111",
	'i':  "00110",
	'j':  "1110101",
	'k':  "101000",
	'l':  "101001",
	'm':  "101010",
	'n':  "00111",
	'o':  "101011",
	'p':  "101100",
	'q':  "1110110",
	'r':  "101101",
	's':  "01000",
	't':  "01001",
	'u':  "101110",
	'v':  "1110111",
	'w':  "1111000",
	'x':  "1111001",
	'y':  "1111010",
	'z':  "1111011",
	'{':  "1111111111111110010",
	'|':  "1111111111101",
	'}':  "1111111111111110011",
	'~':  "1111111111111110100",
	0x7F: "1111111111111110101",
	0x80: "1111111111111111111111111101",
	0x81: "1111111111111110110",
	0x82: "1111111111111110111",
	0x83: "1111111111111111000",
	0x84: "1111111111111111001",
	0x85: "111111111111111111111111111110",
	0x86: "1111111111111111010",
	0x87: "1111111111111111011",
	0x88: "1111111111111111100",
	0x89: "1111111111111111101",
	0x8A: "1111111111111111110",
	0x8B: "111111111111111111111111111111",
	0x8C: "1111111111111111111",
	0x8D: "1111111111111111111110000000",
	0x8E: "1111111111111111111110000001",
	0x8F: "1111111111111111111110000010",
	0x90: "1111111111111111111110000011",
	0x91: "1111111111111111111110000100",
	0x92: "1111111111111111111110000101",
	0x93: "1111111111111111111110000110",
	0x94: "1111111111111111111110000111",
	0x95: "1111111111111111111110001000",
	0x96: "1111111111111111111110001001",
	0x97: "1111111111111111111110001010",
	0x98: "1111111111111111111110001011",
	0x99: "1111111111111111111110001100",
	0x9A: "1111111111111111111110001101",
	0x9B: "1111111111111111111110001110",
	0x9C: "1111111111111111111110001111",
	0x9D: "1111111111111111111110010000",
	0x9E: "1111111111111111111110010001",
	0x9F: "1111111111111111111110010010",
	0xA0: "1111111111111111111111111111110",
}

func HuffmanEncodeString(s string) []byte {
	seq := ""
	for _, char := range s {
		binStr, found := huffmanCodes[char]
		if !found {
			binStr = strconv.FormatInt(int64(char), 2)
		}
		seq += binStr
	}

	paddingLength := (8 - (len(seq) % 8)) % 8
	padding := ""
	for i := 0; i < paddingLength; i++ {
		padding += "1"
	}
	seq += padding

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
	{name: ":status:", value: "204"},
	{name: ":status:", value: "206"},
	{name: ":status:", value: "304"},
	{name: ":status:", value: "400"},
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
	if huffmanEncoded {
		encoded := HuffmanEncodeString(s)
		length := encodeInteger(uint64(len(encoded)), 7)
		length[0] |= 128
		return append(length, encoded...)
	}

	length := encodeInteger(uint64(len(s)), 7)
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
	dynamicTable []HeaderField
}

func NewHPackEncoder() HPackEncoder {
	return &hPackEncoder{
		dynamicTable: []HeaderField{},
	}
}

type headerFieldWithEncodingParams struct {
	headerField      HeaderField
	isHuffmanEncoded bool
	indexed          bool
	newName          bool
}

func (h *hPackEncoder) encodeHeaderField(hf headerFieldWithEncodingParams) []byte {
	indexedHeaderField := 0
	indexedHeaderName := 0
	for index, item := range staticTable {
		if hf.headerField.name == item.name {
			if hf.headerField.value == item.value {
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
	if indexedHeaderName != 0 && hf.indexed && !hf.newName {
		bytes := encodeInteger(uint64(indexedHeaderName), 6)
		bytes[0] |= 0x40

		return append(bytes, encodeStringLiteral(hf.headerField.value, hf.isHuffmanEncoded)...)
	}
	// Literal Header Field with Incremental Indexing -- New Name
	if indexedHeaderName != 0 && hf.indexed && hf.newName {
		bytes := []byte{0x40}

		bytes = append(bytes, encodeStringLiteral(hf.headerField.name, hf.isHuffmanEncoded)...)
		return append(bytes, encodeStringLiteral(hf.headerField.value, hf.isHuffmanEncoded)...)
	}

	// Literal Header Field without Indexing -- Indexed name
	if indexedHeaderName != 0 && !hf.indexed {
		bytes := encodeInteger(uint64(indexedHeaderName), 4)
		bytes[0] &= 0x0f
		return append(bytes, encodeStringLiteral(hf.headerField.value, hf.isHuffmanEncoded)...)
	}

	fmt.Printf("%s: %s\n", hf.headerField.name, hf.headerField.value)
	panic("not implemented yet")
}

func (h *hPackEncoder) getParamsForHeaderField(headerField HeaderField) headerFieldWithEncodingParams {
	// Just a place holder to implement actual function later.
	hfWithParams := headerFieldWithEncodingParams{
		headerField:      headerField,
		isHuffmanEncoded: true,
		indexed:          false,
		newName:          false,
	}
	if headerField.value == "localhost" || headerField.name == "user-agent" || headerField.name == "accept" {
		hfWithParams.indexed = true
	}
	if headerField.value == "*/*" {
		hfWithParams.isHuffmanEncoded = false
	}
	return hfWithParams
}

func (h *hPackEncoder) Encode(writer io.Writer, headerFields []HeaderField) (int, error) {
	// takes an ordered header list and encode it
	total_bytes := 0
	for _, headerField := range headerFields {
		n, err := writer.Write(h.encodeHeaderField(h.getParamsForHeaderField(headerField)))
		if err != nil {
			return total_bytes, err
		}
		total_bytes += n
	}

	return total_bytes, nil
}
