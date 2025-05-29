package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrProtocol = errors.New("protocol error")
)

type (
	FrameType    uint8
	FlagType     uint8
	SettingParam uint16
)

const (
	DataFrameType         FrameType = 0x00
	HeaderFrameType       FrameType = 0x01
	SettingFrameType      FrameType = 0x04
	WindowUpdateFrameType FrameType = 0x08

	UnsetFlag     FlagType = 0x00
	AckFlag       FlagType = 0x01
	EndStreamFlag FlagType = 0x01
	EndHeaderFlag FlagType = 0x04
	PaddedFlag    FlagType = 0x08
	PriorityFlag  FlagType = 0x20

	SettingsHeaderTableSize      SettingParam = 1 // 4,096
	SettingsEnablePush           SettingParam = 2 // true
	SettingsMaxConcurrentStreams SettingParam = 3 // no less than 100
	SettingsInitialWindowSize    SettingParam = 4 // 65,535
	SettingsMaxFrameSize         SettingParam = 5 // 16,384
	SettingsMaxHeaderListSize    SettingParam = 6 // unlimited
)

/*
All frames begin with a fixed 9-octet header followed by a variable-
length payload.

	+-----------------------------------------------+
	|                 Length (24)                   |
	+---------------+---------------+---------------+
	|   Type (8)    |   Flags (8)   |
	+-+-------------+---------------+-------------------------------+
	|R|                 Stream Identifier (31)                      |
	+=+=============================================================+
	|                   Frame Payload (0...)                      ...
	+---------------------------------------------------------------+

	                      Figure 1: Frame Layout
*/
type Frame struct {
	Type     FrameType
	Flags    FlagType
	StreamID uint32
	Data     any
}

/*
The payload of a SETTINGS frame consists of zero or more parameters,
each consisting of an unsigned 16-bit setting identifier and an
unsigned 32-bit value.

	+-------------------------------+
	|       Identifier (16)         |
	+-------------------------------+-------------------------------+
	|                        Value (32)                             |
	+---------------------------------------------------------------+

	                     Figure 10: Setting Format
*/
type SettingFrame struct {
	Params map[SettingParam]uint32
}

/*
Headers frame structure

	+---------------+
	|Pad Length? (8)|
	+-+-------------+-----------------------------------------------+
	|E|                 Stream Dependency? (31)                     |
	+-+-------------+-----------------------------------------------+
	|  Weight? (8)  |
	+-+-------------+-----------------------------------------------+
	|                   Header Block Fragment (*)                 ...
	+---------------------------------------------------------------+
	|                           Padding (*)                       ...
	+---------------------------------------------------------------+

	Figure 7: HEADERS Frame Payload
*/
type HeaderFrame struct {
	StreamDependency uint32
	PaddingLength    uint8
	Weight           uint8

	HeaderFields []HeaderField
}

/*
Window update frame structure

		+-+-------------------------------------------------------------+
	    |R|              Window Size Increment (31)                     |
	    +-+-------------------------------------------------------------+
*/
type WindowUpdateFrame struct {
	WindowSizeIncrement uint32
}

/*
DATA frame structure
    +---------------+
    |Pad Length? (8)|
    +---------------+-----------------------------------------------+
    |                            Data (*)                         ...
    +---------------------------------------------------------------+
    |                           Padding (*)                       ...
    +---------------------------------------------------------------+
*/

type DataFrame struct {
	PadLength uint8
	Data      []byte
}

type frameHandler struct {
	encoder HPackEncoder
	decoder HPackDecoder
}

func NewFrameHandler() *frameHandler {
	return &frameHandler{
		encoder: NewHPackEncoder(),
		decoder: NewHPackDecoder(),
	}
}

func (h *frameHandler) Encode(writer io.Writer, frame Frame) (int, error) {
	switch frame.Type {
	case SettingFrameType:
		settingFrame, ok := frame.Data.(SettingFrame)
		if !ok {
			return 0, fmt.Errorf("invalid data")
		}

		length := len(settingFrame.Params) * 6

		packet := make([]byte, 9+length)
		// Length (24)
		packet[0] = byte((length >> 16) & 0xFF)
		packet[1] = byte((length >> 8) & 0xFF)
		packet[2] = byte(length & 0xFF)
		packet[3] = byte(SettingFrameType)                     // Type (8)
		packet[4] = byte(frame.Flags)                          // Flags (8)
		binary.BigEndian.PutUint32(packet[5:], frame.StreamID) // StreamID (32)

		index := 9
		for identifier, value := range settingFrame.Params {
			binary.BigEndian.PutUint16(packet[index:], uint16(identifier)) // 9-11
			binary.BigEndian.PutUint32(packet[index+2:], value)            // 11-15

			index += 6
		}

		return writer.Write(packet)
	case HeaderFrameType:
		headerFrame, ok := frame.Data.(HeaderFrame)
		if !ok {
			return 0, fmt.Errorf("invalid data")
		}

		packet := make([]byte, 9)
		packet[3] = byte(HeaderFrameType)                      // Type (8)
		packet[4] = byte(frame.Flags)                          // Flags (8)
		binary.BigEndian.PutUint32(packet[5:], frame.StreamID) // StreamID (32)

		if (frame.Flags & PaddedFlag) != UnsetFlag {
			packet = append(packet, headerFrame.PaddingLength)
		}
		if (frame.Flags & PriorityFlag) != UnsetFlag {
			packet = binary.BigEndian.AppendUint32(packet, headerFrame.StreamDependency)
			packet = append(packet, headerFrame.Weight)
		}

		buf := bytes.Buffer{}
		_, err := h.encoder.Encode(&buf, headerFrame.HeaderFields)
		if err != nil {
			return 0, err
		}

		packet = append(packet, buf.Bytes()...)
		length := len(packet) - 9
		packet[0] = byte((length >> 16) & 0xFF)
		packet[1] = byte((length >> 8) & 0xFF)
		packet[2] = byte(length & 0xFF)

		return writer.Write(packet)
	case WindowUpdateFrameType:
		packet := make([]byte, 9+4)
		packet[2] = 4 // Length
		packet[3] = byte(WindowUpdateFrameType)
		binary.BigEndian.PutUint32(packet[5:9], frame.StreamID)
		binary.BigEndian.PutUint32(packet[9:], frame.Data.(WindowUpdateFrame).WindowSizeIncrement)

		return writer.Write(packet)
	case DataFrameType:
		packetLength := 0
		packet := make([]byte, 9)
		packet[3] = byte(DataFrameType)                        // Type (8)
		packet[4] = byte(frame.Flags)                          // Flags (8)
		binary.BigEndian.PutUint32(packet[5:], frame.StreamID) // StreamID (32)

		dataFrame := frame.Data.(DataFrame)
		if (frame.Flags & PaddedFlag) != 0 {
			packet = append(packet, byte(dataFrame.PadLength))
			packetLength += 1
		}
		packetLength += len(dataFrame.Data)
		packet = append(packet, dataFrame.Data...)

		if (frame.Flags & PaddedFlag) != 0 {
			paddingData := make([]byte, dataFrame.PadLength)
			_, err := rand.Read(paddingData)
			if err != nil {
				return 0, err
			}
			packet = append(packet, paddingData...)
		}
		packet[0] = byte((packetLength >> 16) & 0xFF)
		packet[1] = byte((packetLength >> 8) & 0xFF)
		packet[2] = byte(packetLength & 0xFF)

		return writer.Write(packet)
	default:
		return 0, fmt.Errorf("invalid frame type")
	}
}

func (h *frameHandler) Decode(reader io.Reader, frame *Frame) error {
	packet := make([]byte, 9)
	n, err := reader.Read(packet[:9])
	if err != nil {
		return err
	}
	if n != 9 {
		return fmt.Errorf("invalid packet header")
	}

	frameLength := uint32(uint32(packet[0])<<16) + (uint32(packet[1]) << 8) + uint32(packet[2])
	frame.Type = FrameType(packet[3])
	frame.StreamID = binary.BigEndian.Uint32(packet[5:9])
	frame.Flags = FlagType(packet[4])
	// Resize packet size
	packet = make([]byte, int(frameLength))
	_, err = reader.Read(packet)
	if err != nil {
		return err
	}

	switch frame.Type {
	case SettingFrameType:
		settingFrame := SettingFrame{
			Params: map[SettingParam]uint32{},
		}
		for i := 0; i < int(frameLength); i += 6 {
			settingFrame.Params[SettingParam(binary.BigEndian.Uint16(packet[i:i+2]))] = binary.BigEndian.Uint32(packet[i+2 : i+6])
		}
		frame.Data = settingFrame
		return nil
	case HeaderFrameType:
		headerFrame := HeaderFrame{}
		base := 0
		if frame.Flags&PaddedFlag != UnsetFlag {
			headerFrame.PaddingLength = uint8(packet[0])
			base++
		}
		if frame.Flags&PriorityFlag != UnsetFlag {
			headerFrame.StreamDependency = binary.BigEndian.Uint32(packet[base:4])
			headerFrame.Weight = uint8(packet[base+4])
			base += 5
		}

		if err := h.decoder.Decode(bytes.NewBuffer(packet[base:]), &headerFrame.HeaderFields); err != nil {
			return err
		}

		// if headerFrame.PaddingLength != 0 {
		// 	n, err := reader.Read(packet[:headerFrame.PaddingLength])
		// 	if err != nil {
		// 		return err
		// 	}
		// 	if n != int(headerFrame.PaddingLength) {
		// 		return fmt.Errorf("not enough padding")
		// 	}
		// }

		frame.Data = headerFrame
		return nil
	case WindowUpdateFrameType:
		windowUpdateFrame := WindowUpdateFrame{
			WindowSizeIncrement: binary.BigEndian.Uint32(packet[:4]),
		}

		frame.Data = windowUpdateFrame
		return nil
	case DataFrameType:
		dataFrame := DataFrame{}
		if (frame.Flags & PaddedFlag) != 0 {
			dataFrame.PadLength = uint8(packet[0])
			dataFrame.Data = append(dataFrame.Data, packet[1:frameLength-uint32(dataFrame.PadLength)-1]...)
		}
		dataFrame.Data = append(dataFrame.Data, packet...)
		frame.Data = dataFrame
		return nil
	default:
		return fmt.Errorf("unknown frame type")
	}
}
