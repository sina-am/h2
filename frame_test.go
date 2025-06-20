package main

import (
	"bytes"
	"testing"
)

func TestEncodeSettingFrame(t *testing.T) {
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
		Flags:    UnsetFlag,
		Data: SettingFrame{
			Params: map[SettingParam]uint32{
				SettingsMaxConcurrentStreams: 100,
				SettingsInitialWindowSize:    33554432,
				SettingsEnablePush:           0,
			},
		},
	}

	var buf = bytes.Buffer{}
	handler := NewFrameHandler()
	_, err := handler.Encode(&buf, frame)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Errorf("expected data: %v got %v", expected, buf.Bytes())
	}
}
func TestDecodeSettingFrame(t *testing.T) {
	expectedFrame := Frame{
		Type:     SettingFrameType,
		StreamID: 0,
		Flags:    UnsetFlag,
		Data: SettingFrame{
			Params: map[SettingParam]uint32{
				SettingsMaxConcurrentStreams: 100,
				SettingsInitialWindowSize:    33554432,
				SettingsEnablePush:           0,
			},
		},
	}
	rawFrame := []byte{
		0x00, 0x00, 0x12, 0x04,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x03, 0x00,
		0x00, 0x00, 0x64, 0x00,
		0x04, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x02, 0x00,
		0x00, 0x00, 0x00,
	}

	frame := Frame{}
	var buf = bytes.NewBuffer(rawFrame)
	handler := NewFrameHandler()
	err := handler.Decode(buf, &frame)
	if err != nil {
		t.Error(err)
	}

	if frame.Type != expectedFrame.Type || frame.StreamID != expectedFrame.StreamID {
		t.Errorf("not expected")
	}

	settingFrame := frame.Data.(SettingFrame)
	expectedHeaderFrame := expectedFrame.Data.(SettingFrame)

	if frame.Flags != expectedFrame.Flags {
		t.Errorf("excepted flags: %d got %d", expectedFrame.Flags, frame.Flags)
	}
	for key := range settingFrame.Params {
		if settingFrame.Params[key] != expectedHeaderFrame.Params[key] {
			t.Errorf("expected params: %d got %d", expectedHeaderFrame.Params[key], settingFrame.Params[key])
		}
	}
}

func TestEncodeHeaderFrame(t *testing.T) {
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
		Flags:    EndHeaderFlag | EndStreamFlag,
		Data: HeaderFrame{
			StreamDependency: 0,
			PaddingLength:    0,
			Weight:           0,

			HeaderFields: []HeaderField{
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
	_, err := handler.Encode(&buf, frame)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Errorf("expected:\n%v\ngot:\n%v\n", expected, buf.Bytes())
		t.Error("wrong")
	}
}
func TestDecodeHeaderFrame(t *testing.T) {
	expectedFrame := Frame{
		Type:     HeaderFrameType,
		StreamID: 1,
		Flags:    EndStreamFlag | EndHeaderFlag,
		Data: HeaderFrame{
			StreamDependency: 0,
			PaddingLength:    0,
			Weight:           0,

			HeaderFields: []HeaderField{
				{name: ":method:", value: "GET"},
				{name: ":path:", value: "/"},
				{name: ":scheme:", value: "https"},
				{name: ":authority:", value: "localhost"},
				{name: "user-agent", value: "curl/7.85.0"},
				{name: "accept", value: "*/*"},
			},
		},
	}

	rawFrame := []byte{
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

	frame := Frame{}
	var buf = bytes.NewBuffer(rawFrame)
	handler := NewFrameHandler()
	err := handler.Decode(buf, &frame)
	if err != nil {
		t.Error(err)
	}

	if frame.Type != expectedFrame.Type || frame.StreamID != expectedFrame.StreamID {
		t.Errorf("not expected")
	}

	headerFrame := frame.Data.(HeaderFrame)
	expectedHeaderFrame := expectedFrame.Data.(HeaderFrame)

	if frame.Flags != expectedFrame.Flags {
		t.Errorf("excepted flags: %d got %d", expectedFrame.Flags, frame.Flags)
	}

	if headerFrame.StreamDependency != expectedHeaderFrame.StreamDependency {
		t.Errorf("excepted StreamDependency flag: %d got %d", expectedHeaderFrame.StreamDependency, headerFrame.StreamDependency)
	}
	if headerFrame.PaddingLength != expectedHeaderFrame.PaddingLength {
		t.Errorf("excepted padding length: %d got %d", expectedHeaderFrame.PaddingLength, headerFrame.PaddingLength)
	}
	if headerFrame.Weight != expectedHeaderFrame.Weight {
		t.Errorf("excepted weight: %d got %d", expectedHeaderFrame.Weight, headerFrame.Weight)
	}
	for i := 0; i < len(headerFrame.HeaderFields); i++ {
		if headerFrame.HeaderFields[i].name != expectedHeaderFrame.HeaderFields[i].name {
			t.Errorf("expected header field name: %s got %s", expectedHeaderFrame.HeaderFields[i].name, headerFrame.HeaderFields[i].name)
		}
		if headerFrame.HeaderFields[i].value != expectedHeaderFrame.HeaderFields[i].value {
			t.Errorf("expected header field value: %s got %s", expectedHeaderFrame.HeaderFields[i].value, headerFrame.HeaderFields[i].value)
		}
	}
}

func TestDecodeWindowUpdateFrame(t *testing.T) {
	expectedRaw := []byte{
		0x00, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x7f, 0xff, 0x00, 0x00,
	}

	frame := Frame{}
	var buf = bytes.NewBuffer(expectedRaw)
	handler := NewFrameHandler()
	err := handler.Decode(buf, &frame)
	if err != nil {
		t.Error(err)
	}
	windowUpdateFrame := frame.Data.(WindowUpdateFrame)

	if windowUpdateFrame.WindowSizeIncrement != 2147418112 {
		t.Errorf("expected window size increment: %d got %d", 2147418112, windowUpdateFrame.WindowSizeIncrement)
	}
}
func TestEncodeWindowUpdateFrame(t *testing.T) {
	expectedRaw := []byte{
		0x00, 0x00, 0x04, 0x08, 0x00, 0x00, 0x00,
		0x00, 0xa1, 0x7f, 0xff, 0x00, 0x00,
	}

	frame := Frame{
		Type:     WindowUpdateFrameType,
		StreamID: 0xa1,
		Data: WindowUpdateFrame{
			WindowSizeIncrement: 2147418112,
		},
	}

	buf := bytes.Buffer{}
	handler := NewFrameHandler()
	_, err := handler.Encode(&buf, frame)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expectedRaw, buf.Bytes()) {
		t.Error("encoded data is not equal to expected raw data")
	}
}
func TestEncodeDataFrame(t *testing.T) {
	expected := []byte{
		0x00, 0x00, 0x0b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72,
		0x6c, 0x64,
	}
	frame := Frame{
		Type:     DataFrameType,
		StreamID: 1,
		Flags:    EndStreamFlag,
		Data: DataFrame{
			PadLength: 0,
			Data:      []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64},
		},
	}

	var buf = bytes.Buffer{}
	handler := NewFrameHandler()
	_, err := handler.Encode(&buf, frame)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Errorf("expected data: %v got %v", expected, buf.Bytes())
	}
}
func TestDecodeDataFrame(t *testing.T) {
	expectedFrame := Frame{
		Type:     DataFrameType,
		StreamID: 1,
		Flags:    EndStreamFlag,
		Data: DataFrame{
			PadLength: 0,
			Data:      []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64},
		},
	}

	raw := []byte{
		0x00, 0x00, 0x0b, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72,
		0x6c, 0x64,
	}

	handler := NewFrameHandler()
	frame := Frame{}
	err := handler.Decode(bytes.NewBuffer(raw), &frame)
	if err != nil {
		t.Error(err)
	}

	if frame.Type != expectedFrame.Type {
		t.Errorf("expected frame type: %d got %d", expectedFrame.Type, frame.Type)
	}
	if frame.StreamID != expectedFrame.StreamID {
		t.Errorf("expected frame stream ID: %d got %d", expectedFrame.StreamID, frame.StreamID)
	}
	if frame.Flags != expectedFrame.Flags {
		t.Errorf("expected frame flags: %d got %d", expectedFrame.Flags, frame.Flags)
	}
	if !bytes.Equal(frame.Data.(DataFrame).Data, expectedFrame.Data.(DataFrame).Data) {
		t.Errorf("expected data: %v got %v", expectedFrame.Data.(DataFrame).Data, frame.Data.(DataFrame).Data)
	}
}
