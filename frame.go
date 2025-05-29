package main

import (
	"bytes"
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

	UnsetFlag           FlagType = 0x0
	SettingAckFlag      FlagType = 0x1
	HeaderEndStreamFlag FlagType = 0x1
	HeaderEndHeaderFlag FlagType = 0x4
	HeaderPaddingFlag   FlagType = 0x8
	HeaderPriorityFlag  FlagType = 0x20

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
	AckFlag FlagType
	Params  map[SettingParam]uint32
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
	EndStreamFlag FlagType
	EndHeaderFlag FlagType
	PaddingFlag   FlagType
	PriorityFlag  FlagType

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
		packet[4] = byte(settingFrame.AckFlag)                 // Flags (8)
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
		packet[3] = byte(HeaderFrameType) // Type (8)
		packet[4] = byte(
			headerFrame.EndStreamFlag |
				headerFrame.EndHeaderFlag |
				headerFrame.PaddingFlag |
				headerFrame.PriorityFlag,
		) // Flags (8)
		binary.BigEndian.PutUint32(packet[5:], frame.StreamID) // StreamID (32)

		if headerFrame.PaddingFlag != UnsetFlag {
			packet = append(packet, headerFrame.PaddingLength)
		}
		if headerFrame.PriorityFlag != UnsetFlag {
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

	// Resize packet size
	packet = make([]byte, int(frameLength))

	switch frame.Type {
	case SettingFrameType:
		settingFrame := SettingFrame{
			AckFlag: FlagType(packet[4]) & SettingAckFlag,
			Params:  map[SettingParam]uint32{},
		}
		for i := 0; i < int(frameLength)/(6); i++ {
			_, err := reader.Read(packet[:6])
			if err != nil {
				return err
			}
			settingFrame.Params[SettingParam(binary.BigEndian.Uint16(packet[:2]))] = binary.BigEndian.Uint32(packet[2:6])
		}
		frame.Data = settingFrame
		return nil
	case HeaderFrameType:
		headerFrame := HeaderFrame{
			EndStreamFlag: FlagType(packet[4]) & HeaderEndStreamFlag,
			EndHeaderFlag: FlagType(packet[4]) & HeaderEndHeaderFlag,
			PaddingFlag:   FlagType(packet[4]) & HeaderPaddingFlag,
			PriorityFlag:  FlagType(packet[4]) & HeaderPriorityFlag,
		}
		if headerFrame.PaddingFlag == HeaderPaddingFlag {
			_, err := reader.Read(packet[:1])
			if err != nil {
				return err
			}
			headerFrame.PaddingLength = uint8(packet[0])
		}
		if headerFrame.PriorityFlag == HeaderPriorityFlag {
			_, err := reader.Read(packet[:5])
			if err != nil {
				return err
			}
			headerFrame.StreamDependency = binary.BigEndian.Uint32(packet[:4])
			headerFrame.Weight = uint8(packet[4])
		}

		if err := h.decoder.Decode(reader, &headerFrame.HeaderFields); err != nil {
			return err
		}

		if headerFrame.PaddingLength != 0 {
			n, err := reader.Read(packet[:headerFrame.PaddingLength])
			if err != nil {
				return err
			}
			if n != int(headerFrame.PaddingLength) {
				return fmt.Errorf("not enough padding")
			}
		}

		frame.Data = headerFrame
		return nil
	case WindowUpdateFrameType:
		_, err := reader.Read(packet)
		if err != nil {
			return err
		}
		windowUpdateFrame := WindowUpdateFrame{
			WindowSizeIncrement: binary.BigEndian.Uint32(packet[:4]),
		}

		frame.Data = windowUpdateFrame
		return nil
	default:
		return fmt.Errorf("unknown frame type")
	}
}
