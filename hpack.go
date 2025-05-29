package main

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrDecodingNumber = errors.New("invalid byte for numeric representation")
)

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
		encoded := HuffmanEncode(s)
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
		return HuffmanDecode(b), nil
	}
	return string(b), nil
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

type HPackDecoder interface {
	Decode(reader io.Reader, headerFields *[]HeaderField) error
}

type hPackDecoder struct {
	dynamicTable []HeaderField
	table        []HeaderField
}

func NewHPackDecoder() HPackDecoder {
	return &hPackDecoder{
		dynamicTable: []HeaderField{},
		table:        staticTable,
	}
}

func (h *hPackDecoder) Decode(reader io.Reader, headerFields *[]HeaderField) error {
	var (
		n     int = 0
		err   error
		value string
		field string
	)

	for {
		bytes := [100]byte{}
		n, err = reader.Read(bytes[:1])
		if err != nil {
			if n == 0 {
				break
			}
			return err
		}

		if bytes[0] >= 0x80 {
			// Indexed Header Field Representation
			bytes[0] -= 0x80
			index, err := decodeInteger(bytes[:1], 7)
			if err != nil {
				return err
			}

			*headerFields = append(*headerFields, h.table[index])
		} else if bytes[0] >= 0x40 {
			// Literal Header Field with Incremental Indexing -- Indexed name
			bytes[0] -= 0x40
			index, err := decodeInteger(bytes[:1], 6)
			if err != nil {
				return err
			}

			// Decode value
			_, err = reader.Read(bytes[:1])
			if err != nil {
				return err
			}

			huffmanEncoded := bytes[0] > 0x80
			if huffmanEncoded {
				bytes[0] -= 0x80
			}

			length, err := decodeInteger(bytes[:1], 7)
			if err != nil {
				return err
			}

			n, err = reader.Read(bytes[:length])
			if err != nil {
				return err
			}
			value, err = decodeStringLiteral(bytes[:n], huffmanEncoded)
			if err != nil {
				return err
			}
			newHeader := HeaderField{name: h.table[index].name, value: value}
			*headerFields = append(*headerFields, newHeader)
			h.table = append(h.table, newHeader)
		} else if bytes[0] == 0 {
			// Literal Header Field without Indexing -- New Name

			// Read the header field
			_, err = reader.Read(bytes[:1])
			if err != nil {
				return err
			}

			huffmanEncoded := bytes[0] > 0x80
			if huffmanEncoded {
				bytes[0] -= 0x80
			}
			length, err := decodeInteger(bytes[:1], 6)
			if err != nil {
				return err
			}

			n, err = reader.Read(bytes[:length])
			if err != nil {
				return err
			}
			field, err = decodeStringLiteral(bytes[:n], huffmanEncoded)
			if err != nil {
				return err
			}
			// Read the header value
			_, err = reader.Read(bytes[:1])
			if err != nil {
				return err
			}

			huffmanEncoded = bytes[0] > 0x80
			if huffmanEncoded {
				bytes[0] -= 0x80
			}
			length, err = decodeInteger(bytes[:1], 6)
			if err != nil {
				return err
			}

			n, err = reader.Read(bytes[:length])
			if err != nil {
				return err
			}
			value, err = decodeStringLiteral(bytes[:n], huffmanEncoded)
			if err != nil {
				return err
			}

			*headerFields = append(*headerFields, HeaderField{name: field, value: value})
		}

		if n == 0 {
			break
		}
	}
	return nil
}
