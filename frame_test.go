package main

import (
	"bytes"
	"testing"
)

func TestSettingFrame(t *testing.T) {
	expected := []byte{
		0x00, 0x00, 0x12, 0x04,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x03, 0x00,
		0x00, 0x00, 0x64, 0x00,
		0x04, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x02, 0x00,
		0x00, 0x00, 0x00,
	}
	frame := Frame{
		Type:     SettingFrameType,
		StreamID: 0,
		Data: SettingFrame{
			AckFlag: UnsetFlag,
			Params: map[SettingParam]uint32{
				SettingsMaxConcurrentStreams: 100,
				SettingsInitialWindowSize:    33554432,
				SettingsEnablePush:           0,
			},
		},
	}

	var buf = bytes.Buffer{}
	handler := NewFrameHandler()
	_, err := handler.Serialize(&buf, frame)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Error("wrong")
	}
}

func TestHeaderFrame(t *testing.T) {
	expected := []byte{
		0x00, 0x00, 0x1a, 0x01,
		0x05, 0x00, 0x00, 0x00,
		0x01, 0x82, 0x84, 0x87,
		0x41, 0x86, 0xa0, 0xe4,
		0x1d, 0x13, 0x9d, 0x09,
		0x7a, 0x88, 0x25, 0xb6,
		0x50, 0xc3, 0xab, 0xbc,
		0xda, 0xe0, 0x53, 0x03,
		0x2a, 0x2f, 0x2a,
	}
	frame := Frame{
		Type:     HeaderFrameType,
		StreamID: 1,
		Data: HeaderFrame{
			EndStreamFlag: HeaderEndStreamFlag,
			EndHeaderFlag: HeaderEndHeaderFlag,
			PaddingFlag:   UnsetFlag,
			PriorityFlag:  UnsetFlag,

			StreamDependency: 0,
			PaddingLength:    0,
			Weight:           0,

			headerFields: []HeaderField{
				{name: ":method:", value: "GET"},
				{name: ":path:", value: "/"},
				{name: ":scheme:", value: "https"},
				{name: ":authority:", value: "localhost"},
				{name: "user-agent", value: "curl/7.85.0"},
				{name: "accept", value: "*/*"},
			},
		},
	}

	var buf = bytes.Buffer{}
	handler := NewFrameHandler()
	_, err := handler.Serialize(&buf, frame)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Errorf("expected:\n%v\ngot:\n%v\n", expected, buf.Bytes())
		t.Error("wrong")
	}
}
